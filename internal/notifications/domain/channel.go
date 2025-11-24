package domain

// Channel represents the delivery channel for a notification.
type Channel string

const (
	ChannelEmail Channel = "email"
	ChannelSMS   Channel = "sms"
	ChannelPush  Channel = "push"
)

// String returns the string representation.
func (c Channel) String() string {
	return string(c)
}

// IsValid checks if the channel is valid.
func (c Channel) IsValid() bool {
	switch c {
	case ChannelEmail, ChannelSMS, ChannelPush:
		return true
	default:
		return false
	}
}
