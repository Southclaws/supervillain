package state

import "encoding/json"

type State int

const (
	StateUnknown    State = 0
	StateProcessing State = 1
	StateSuccess    State = 2
	StateFailed     State = 3
)

const (
	StateStringUnknown    = "unknown"
	StateStringProcessing = "processing"
	StateStringSuccess    = "success"
	StateStringFailed     = "failed"
)

func (s State) String() string {
	switch s {
	case StateProcessing:
		return StateStringProcessing
	case StateSuccess:
		return StateStringSuccess
	case StateFailed:
		return StateStringFailed
	default:
		return StateStringUnknown
	}
}

func ParseState(s string) State {
	switch s {
	case StateStringProcessing:
		return StateProcessing
	case StateStringSuccess:
		return StateSuccess
	case StateStringFailed:
		return StateFailed
	default:
		return StateUnknown
	}
}

func (s State) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s State) ZodSchema() string {
	return "z.string()"
}
