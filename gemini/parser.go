package gemini

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type rawEnvelope struct {
	Type string `json:"type"`
	Role string `json:"role,omitempty"`
}

// ParseMessage parses a raw JSON line into a typed Message.
// Unknown types are returned as *UnknownMessage (never an error).
func ParseMessage(data []byte) (Message, error) {
	var env rawEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, &ProtocolError{
			Message: fmt.Sprintf("invalid JSON: %v", err),
			Raw:     bytes.Clone(data),
		}
	}

	switch env.Type {
	case "init":
		return decode[InitMessage](data)
	case "message":
		switch env.Role {
		case "assistant":
			return decode[AssistantMessage](data)
		case "user":
			return decode[UserMessage](data)
		default:
			return &UnknownMessage{RawType: "message/" + env.Role, Raw: bytes.Clone(data)}, nil
		}
	case "tool_use":
		return decode[ToolUseMessage](data)
	case "tool_result":
		return decode[ToolResultMessage](data)
	case "result":
		return decode[ResultMessage](data)
	case "error":
		return decode[ErrorMessage](data)
	default:
		return &UnknownMessage{RawType: env.Type, Raw: bytes.Clone(data)}, nil
	}
}

// decode unmarshals JSON into a message type and returns it as a Message.
func decode[T any, PT interface {
	*T
	Message
}](data []byte) (Message, error) {
	var msg T
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, &ProtocolError{
			Message: fmt.Sprintf("parsing %T: %v", msg, err),
			Raw:     bytes.Clone(data),
		}
	}
	return PT(&msg), nil
}
