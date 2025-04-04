package smtp

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"
)

const boundary = "mail-boundary"

// Email struct contains the parameters necessary to create an email.
type Email struct {
	to      []string
	from    string
	subject string
	body    string

	attachments []attachment
}

type attachment struct {
	name string
	data []byte
}

// NewEmail is the Email constructor.
func NewEmail(to []string, from, subject, body string) *Email {
	return &Email{
		to:      to,
		from:    from,
		subject: subject,
		body:    body,
	}
}

// Attach attaches the binary data as a file with given name to the email.
func (e *Email) Attach(name string, data []byte) {
	e.attachments = append(e.attachments, attachment{
		name: name,
		data: data,
	})
}

// Build returns the email data ready for sending.
// The email is building according to RFC 2045 (MIME).
//
//nolint:revive // Impossible WriteString errors.
func (e *Email) Build() []byte {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("From: %s\r\n", e.from))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(e.to, ";")))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", e.subject))

	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n", boundary))
	buf.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))
	buf.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	buf.WriteString("\r\n" + e.body)

	for _, atch := range e.attachments {
		buf.WriteString(fmt.Sprintf("\r\n\r\n--%s\r\n", boundary))
		buf.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
		buf.WriteString("Content-Transfer-Encoding: base64\r\n")
		buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=%s\r\n", atch.name))
		buf.WriteString(fmt.Sprintf("Content-ID: <%s>\r\n\r\n", atch.name))

		b := make([]byte, base64.StdEncoding.EncodedLen(len(atch.data)))
		base64.StdEncoding.Encode(b, atch.data)
		buf.Write(b)
	}

	buf.WriteString(fmt.Sprintf("\r\n\r\n--%s--", boundary))

	return buf.Bytes()
}
