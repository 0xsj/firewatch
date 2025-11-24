package email

// Message represents an email message to be sent.
type Message struct {
	// From is the sender address.
	// If empty, the default sender from config is used.
	From string

	// To is the list of recipient addresses.
	To []string

	// CC is the list of carbon copy addresses.
	CC []string

	// BCC is the list of blind carbon copy addresses.
	BCC []string

	// ReplyTo is the reply-to address.
	ReplyTo string

	// Subject is the email subject line.
	Subject string

	// TextBody is the plain text body.
	TextBody string

	// HTMLBody is the HTML body.
	HTMLBody string

	// Headers contains additional email headers.
	Headers map[string]string
}

// NewMessage creates a new email message with required fields.
func NewMessage(to []string, subject, textBody string) Message {
	return Message{
		To:       to,
		Subject:  subject,
		TextBody: textBody,
	}
}

// WithFrom sets the sender address.
func (m Message) WithFrom(from string) Message {
	m.From = from
	return m
}

// WithHTML sets the HTML body.
func (m Message) WithHTML(html string) Message {
	m.HTMLBody = html
	return m
}

// WithCC sets the CC recipients.
func (m Message) WithCC(cc []string) Message {
	m.CC = cc
	return m
}

// WithBCC sets the BCC recipients.
func (m Message) WithBCC(bcc []string) Message {
	m.BCC = bcc
	return m
}

// WithReplyTo sets the reply-to address.
func (m Message) WithReplyTo(replyTo string) Message {
	m.ReplyTo = replyTo
	return m
}

// WithHeader adds a custom header.
func (m Message) WithHeader(key, value string) Message {
	if m.Headers == nil {
		m.Headers = make(map[string]string)
	}
	m.Headers[key] = value
	return m
}
