// Package smtp wraps the standard package net/smtp,
// provides user-friendly API for sending emails via an SMTP server.
// It allows to create an SMTP client with a permanent TCP connection to reuse it in the future.
// Also provides functions for creating emails according to RFC 2045 (MIME).
// It supports attachments with binary data.
package smtp

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

const (
	extStartTLS = "STARTTLS"
	extAuth     = "AUTH"
)

var (
	ErrUnsupportedAuthExt   = errors.New("extension is unsupported by SMTP-server, extension: " + extAuth)
	ErrFoundCRLF            = errors.New("a line must not contain CR or LF")
	ErrClientNotInitialized = errors.New("client not initialized")
)

// Client is an SMTP client.
type Client struct {
	opts   Options
	auth   smtp.Auth
	client *smtp.Client
}

// Options contains parameters to build a Client.
type Options struct {
	Host     string
	Port     int
	Login    string
	Password string
}

// NewClient is a Client constructor.
func NewClient(opts Options) *Client {
	return &Client{
		opts: opts,
		auth: smtp.PlainAuth("", opts.Login, opts.Password, opts.Host),
	}
}

// Connect connects to the specified SMTP server address
// and authenticates the client with the specified login and password.
func (c *Client) Connect(ctx context.Context) error {
	addr := net.JoinHostPort(c.opts.Host, strconv.Itoa(c.opts.Port))

	dialer := net.Dialer{Timeout: dialTimeout}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to dialer.DialContext, address: %s: %w", addr, err)
	}

	client, err := smtp.NewClient(conn, c.opts.Host)
	if err != nil {
		return fmt.Errorf("failed to smtp.NewClient: %w", err)
	}

	if err = client.Hello(c.opts.Host); err != nil {
		return fmt.Errorf("failed to client.Hello: %w", err)
	}

	if ok, _ := client.Extension(extStartTLS); ok {
		config := &tls.Config{
			ServerName: c.opts.Host,
			MinVersion: tls.VersionTLS12,
		}
		if err = client.StartTLS(config); err != nil {
			return fmt.Errorf("failed to client.StartTLS: %w", err)
		}
	}

	if c.auth != nil {
		if ok, _ := client.Extension(extAuth); !ok {
			return ErrUnsupportedAuthExt
		}

		if err = client.Auth(c.auth); err != nil {
			return fmt.Errorf("failed to client.Auth: %w", err)
		}
	}

	c.client = client

	return nil
}

// EnsureConnected provides connection to the SMTP server.
// This function checks whether the connection was established earlier.
// If so - it executes NOOP command to check that the connection is still available.
// If NOOP returns an error or the connection was not established earlier,
// it tries to connect once.
func (c *Client) EnsureConnected(ctx context.Context) error {
	if c.client != nil {
		if err := c.client.Noop(); err != nil {
			return c.Connect(ctx)
		}
		return nil
	}

	return c.Connect(ctx)
}

const dialTimeout = 30 * time.Second

// Send sends a message with SMTP commands MAIL, RCPT and DATA.
func (c *Client) Send(to []string, from string, msg []byte) error {
	if err := validateLine(from); err != nil {
		return fmt.Errorf("failed to validateLine: %w", err)
	}

	for _, recp := range to {
		if err := validateLine(recp); err != nil {
			return fmt.Errorf("failed to validateLine: %w", err)
		}
	}

	if c.client == nil {
		return ErrClientNotInitialized
	}

	if err := c.client.Mail(from); err != nil {
		return fmt.Errorf("failed to client.Mail: %w", err)
	}

	for _, addr := range to {
		if err := c.client.Rcpt(addr); err != nil {
			return fmt.Errorf("failed to client.Rcpt: %w", err)
		}
	}

	w, err := c.client.Data()
	if err != nil {
		return fmt.Errorf("failed to client.Data: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close message: %w", err)
	}

	return nil
}

func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Quit()
	}
	return nil
}

// validateLine realization from net/smtp: checks to see if a line has CR or LF as per RFC 5321.
func validateLine(line string) error {
	if strings.ContainsAny(line, "\r\n") {
		return ErrFoundCRLF
	}
	return nil
}
