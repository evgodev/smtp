package smtp

import (
	"strings"
	"testing"
)

func TestMail(t *testing.T) {
	type fields struct {
		to          []string
		from        string
		subject     string
		body        string
		attachments []attachment
	}

	const (
		from    = "from@outlook.com"
		subject = "test subject"
		body    = "test body"
	)

	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		{
			name: "no attachments, one recipient",
			fields: fields{
				to:          []string{"to@outlook.com"},
				from:        from,
				subject:     subject,
				body:        body,
				attachments: nil,
			},
			want: []byte(`From: from@outlook.com
To: to@outlook.com
Subject: test subject
MIME-Version: 1.0
Content-Type: multipart/mixed; boundary=mail-boundary

--mail-boundary
Content-Type: text/plain; charset="utf-8"

test body

--mail-boundary--`),
		},
		{
			name: "no attachments, 3 recipients",
			fields: fields{
				to: []string{
					"test-1@outlook.com",
					"test-2@gmail.com",
					"test-3@mail.ru",
				},
				from:        from,
				subject:     subject,
				body:        body,
				attachments: nil,
			},
			want: []byte(`From: from@outlook.com
To: test-1@outlook.com;test-2@gmail.com;test-3@mail.ru
Subject: test subject
MIME-Version: 1.0
Content-Type: multipart/mixed; boundary=mail-boundary

--mail-boundary
Content-Type: text/plain; charset="utf-8"

test body

--mail-boundary--`),
		},
		{
			name: "one attachment",
			fields: fields{
				to: []string{
					"test-1@outlook.com",
					"test-2@gmail.com",
					"test-3@mail.ru",
				},
				from:    from,
				subject: subject,
				body:    body,
				attachments: []attachment{
					{
						name: "attachment_1.txt",
						data: []byte("attachment_1.txt"),
					},
				},
			},
			want: []byte(`From: from@outlook.com
To: test-1@outlook.com;test-2@gmail.com;test-3@mail.ru
Subject: test subject
MIME-Version: 1.0
Content-Type: multipart/mixed; boundary=mail-boundary

--mail-boundary
Content-Type: text/plain; charset="utf-8"

test body

--mail-boundary
Content-Type: text/plain; charset="utf-8"
Content-Transfer-Encoding: base64
Content-Disposition: attachment; filename=attachment_1.txt
Content-ID: <attachment_1.txt>

YXR0YWNobWVudF8xLnR4dA==

--mail-boundary--`),
		},
		{
			name: "two attachments",
			fields: fields{
				to: []string{
					"test-1@outlook.com",
					"test-2@gmail.com",
					"test-3@mail.ru",
				},
				from:    "from@outlook.com",
				subject: "test subject",
				body:    "test body",
				attachments: []attachment{
					{
						name: "attachment_1.txt",
						data: []byte("attachment_1.txt"),
					},
					{
						name: "attachment_2.txt",
						data: []byte("attachment_2.txt"),
					},
				},
			},
			want: []byte(`From: from@outlook.com
To: test-1@outlook.com;test-2@gmail.com;test-3@mail.ru
Subject: test subject
MIME-Version: 1.0
Content-Type: multipart/mixed; boundary=mail-boundary

--mail-boundary
Content-Type: text/plain; charset="utf-8"

test body

--mail-boundary
Content-Type: text/plain; charset="utf-8"
Content-Transfer-Encoding: base64
Content-Disposition: attachment; filename=attachment_1.txt
Content-ID: <attachment_1.txt>

YXR0YWNobWVudF8xLnR4dA==

--mail-boundary
Content-Type: text/plain; charset="utf-8"
Content-Transfer-Encoding: base64
Content-Disposition: attachment; filename=attachment_2.txt
Content-ID: <attachment_2.txt>

YXR0YWNobWVudF8yLnR4dA==

--mail-boundary--`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mail := NewEmail(
				test.fields.to,
				test.fields.from,
				test.fields.subject,
				test.fields.body,
			)

			for _, atch := range test.fields.attachments {
				mail.Attach(atch.name, atch.data)
			}

			gotMailBytes := mail.Build()

			// On Linux and macOS the new line character is \n (LF), on Windows it is \r\n (CRLF).
			// According to RFC 2045 (MIME) 2.1, we must use CRLF as a line break in emails.
			// Our expected mails in test cases on Linux are built using LF only.
			// This is why we will have to replace LF with CRLF for comparison in the test.
			gotMail := string(gotMailBytes)
			wantMailCRLF := strings.ReplaceAll(string(test.want), "\n", "\r\n")

			requireEqual(t, wantMailCRLF, gotMail)
		})
	}
}

func requireEqual(t *testing.T, got, want interface{}) {
	if got != want {
		t.Errorf("got: %v, want: %v", got, want)
	}
}
