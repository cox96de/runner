package handler

import "encoding/json"

// Message is a struct for error message.
type Message struct {
	Message error
}

func (m *Message) MarshalJSON() ([]byte, error) {
	mm := map[string]interface{}{"message": m.Message}
	return json.Marshal(mm)
}
