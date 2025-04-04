# SMTP Package

A user-friendly Go library that wraps the standard `net/smtp` package, providing an improved API for sending emails via
SMTP servers.

## Features

- Simplified SMTP client creation with persistent TCP connections
- RFC 2045 (MIME) compliant email creation
- Support for email attachments with binary data
- Connection pooling and reuse

## Installation

``` bash
go get github.com/yourusername/smtp
```

## Usage

### Creating an SMTP Client

``` go
client := smtp.NewClient(smtp.Options{
    Host:     "smtp.example.com",
    Port:     587,
    Login:    "user@example.com",
    Password: "password",
})

// Connect to the SMTP server
ctx := context.Background()
err := client.Connect(ctx)
if err != nil {
    log.Fatal(err)
}
defer client.Close()
```

### Creating and Sending Emails

``` go
// Create an email
email := smtp.NewEmail(
    []string{"recipient@example.com"},
    "sender@example.com",
    "Hello from Go",
    "This is the email body text.",
)

// Add attachments (optional)
email.Attach("document.txt", []byte("This is the attachment content"))

// Build the email
emailData := email.Build()

// Send the email
err = client.Send(
    []string{"recipient@example.com"},
    "sender@example.com",
    emailData,
)
if err != nil {
    log.Fatal(err)
}
```

### Ensuring Connection Availability

``` go
// Check connection and reconnect if needed
err := client.EnsureConnected(ctx)
if err != nil {
    log.Fatal(err)
}
```

## API Reference

### SMTP Client

| Function                                           | Description                                                    |
|----------------------------------------------------|----------------------------------------------------------------|
| `NewClient(opts Options) *Client`                  | Creates a new SMTP client with the provided options.           |
| `Connect(ctx context.Context) error`               | Establishes a connection to the SMTP server and authenticates. |
| `EnsureConnected(ctx context.Context) error`       | Ensures the client is connected, reconnecting if necessary.    |
| `Send(to []string, from string, msg []byte) error` | Sends an email to the specified recipients.                    |
| `Close() error`                                    | Closes the connection to the SMTP server.                      |

### Email Creation

| Function                                                   | Description                                                                                     |
|------------------------------------------------------------|-------------------------------------------------------------------------------------------------|
| `NewEmail(to []string, from, subject, body string) *Email` | Creates a new email message with the specified recipients, sender, subject, and body.           |
| `Attach(name string, data []byte)`                         | Attaches binary data to the email with the specified filename.                                  |
| `Build() []byte`                                           | Builds the email according to RFC 2045 (MIME) and returns it as a byte slice ready for sending. |
