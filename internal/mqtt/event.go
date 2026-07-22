package mqtt

// Event is the JSON envelope published to EMQX inbox topics.
type Event struct {
	Type      string `json:"type"`
	RequestID string `json:"request_id,omitempty"`
	Payload   any    `json:"payload"`
}

const EventTypeMessageNew = "message.new"
