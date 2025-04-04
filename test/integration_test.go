//go:build integration

package test

import (
	"context"
	"testing"

	"github.com/evgodev/smtp"
)

// opts - correct SMTP connection credentials.
var opts = smtp.Options{
	Host:     "localhost",
	Port:     1025,
	Login:    "integration-test-user",
	Password: "integration-test-password",
}

func TestConnect(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		opts      smtp.Options
		wantError bool
	}{
		{
			name:      "valid credentials",
			opts:      opts,
			wantError: false,
		},
		{
			name: "incorrect host",
			opts: smtp.Options{
				Host:     "incorrect_host",
				Port:     opts.Port,
				Login:    opts.Login,
				Password: opts.Password,
			},
			wantError: true,
		},
		{
			name: "incorrect port",
			opts: smtp.Options{
				Host:     opts.Host,
				Port:     10,
				Login:    opts.Login,
				Password: opts.Password,
			},
			wantError: true,
		},
		{
			name: "incorrect login",
			opts: smtp.Options{
				Host:     opts.Host,
				Port:     opts.Port,
				Login:    "incorrect_login",
				Password: opts.Password,
			},
			wantError: true,
		},
		{
			name: "incorrect password",
			opts: smtp.Options{
				Host:     opts.Host,
				Port:     opts.Port,
				Login:    opts.Login,
				Password: "incorrect_password",
			},
			wantError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			client := smtp.NewClient(test.opts)

			err := client.Connect(context.Background())
			if test.wantError {
				requireError(t, err, "client.Connect(...) must error")
				return
			}
			requireNoError(t, err, "client.Connect(...) unexpected error")

			err = client.Close()
			requireNoError(t, err, "client.Close() unexpected error")
		})
	}
}

func TestEnsureConnected(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cred smtp.Options
		act  func(t *testing.T, c *smtp.Client)
	}{
		{
			name: "ensure connected after connect",
			cred: opts,
			act: func(t *testing.T, c *smtp.Client) {
				t.Helper()
				ctx := context.Background()

				err := c.Connect(ctx)
				requireNoError(t, err, "client.Connect(...) unexpected error")

				err = c.EnsureConnected(ctx)
				requireNoError(t, err, "client.EnsureConnected(...) unexpected error")
			},
		},
		{
			name: "ensure connected without connecting before",
			cred: opts,
			act: func(t *testing.T, c *smtp.Client) {
				t.Helper()
				ctx := context.Background()

				err := c.EnsureConnected(ctx)
				requireNoError(t, err, "unexpected error when connecting for the first time")

				err = c.EnsureConnected(ctx)
				requireNoError(t, err, "unexpected error when already connected")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			client := smtp.NewClient(test.cred)
			t.Cleanup(func() {
				err := client.Close()
				requireNoError(t, err, "client.Close() unexpected error")
			})

			if test.act != nil {
				test.act(t, client)
			}
		})
	}
}

func TestBuildAndSendMail(t *testing.T) {
	type attachment struct {
		name string
		data []byte
	}

	type email struct {
		to          []string
		from        string
		subject     string
		body        string
		attachments []attachment
	}

	tests := []struct {
		name  string
		email email
	}{
		{
			name: "one recipient, no attachments",
			email: email{
				to:          []string{"to@domain.com"},
				from:        "from@domain.com",
				subject:     "one recipient, no attachments",
				body:        "test mail body",
				attachments: nil,
			},
		},
		{
			name: "many recipients, no attachments",
			email: email{
				to: []string{
					"recipient-1@outlook.com",
					"recipient-2@gmail.com",
					"recipient-3@mail.ru",
				},
				from:        "from@domain.com",
				subject:     "many recipients, no attachments",
				body:        "test mail body",
				attachments: nil,
			},
		},
		{
			name: "few attachments",
			email: email{
				to: []string{
					"recipient-1@outlook.com",
					"recipient-2@gmail.com",
					"recipient-3@mail.ru",
				},
				from:    "from@domain.com",
				subject: "many attachments",
				body:    string(loremIpsum),
				attachments: []attachment{
					{
						name: "lorem_ipsum.txt",
						data: loremIpsum,
					},
					{
						name: "lorem_ipsum2.log",
						data: loremIpsum,
					},
					{
						name: "lorem_ipsum",
						data: loremIpsum,
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			client := smtp.NewClient(opts)
			err := client.Connect(context.Background())
			requireNoError(t, err, "client.Connect(...) unexpected error")

			newMail := smtp.NewEmail(
				test.email.to,
				test.email.from,
				test.email.subject,
				test.email.body,
			)

			for _, atch := range test.email.attachments {
				newMail.Attach(atch.name, atch.data)
			}

			err = client.Send(test.email.to, test.email.from, newMail.Build())
			requireNoError(t, err, "client.Send(...) unexpected error")

			err = client.Close()
			requireNoError(t, err, "client.Close() unexpected error")
		})
	}
}

var loremIpsum = []byte(`Lorem ipsum dolor sit amet, consectetur adipiscing 
elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim 
ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip 
ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate 
velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat 
cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est 
laborum. Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do.
Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor 
incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis 
nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. 
Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu 
fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in 
culpa qui officia deserunt mollit anim id est laborum. Lorem ipsum dolor sit amet.`)

func requireNoError(t *testing.T, err error, msg string) {
	if err != nil {
		t.Errorf("%s: %v", msg, err)
	}
}

func requireError(t *testing.T, err error, msg string) {
	if err == nil {
		t.Errorf("%s: %v", msg, err)
	}
}
