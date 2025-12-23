package senders

// Message represents an email to be sent
type Message struct {
	To      string
	Subject string
	Body    string
	IsHTML  bool
}
