package evidence

import (
	"encoding/json"
	"slices"
	"strings"
)

// A user gate declares that a turn's open list cannot advance without the user.
// It is read off the receipt: a gated turn and one that stopped early have the
// same shape.

// UserGateTool is the call that records a gate.
const UserGateTool = "await_user"

// UserGate is a recorded gate: the waiting item, and what the user must supply.
type UserGate struct {
	StepID string
	Need   string
}

// UserGateThisTurn returns the gate recorded this turn. Only a successful call
// counts; a refused one leaves readiness unchanged.
func (l *Ledger) UserGateThisTurn() (UserGate, bool) {
	if l == nil {
		return UserGate{}, false
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, r := range slices.Backward(l.receipts) {
		if !r.Success || r.ToolName != UserGateTool {
			continue
		}
		return decodeUserGate(r.Args), true
	}
	return UserGate{}, false
}

func decodeUserGate(args json.RawMessage) UserGate {
	var payload struct {
		StepID string `json:"step_id"`
		Need   string `json:"need"`
	}
	if len(args) == 0 || json.Unmarshal(args, &payload) != nil {
		return UserGate{}
	}
	return UserGate{StepID: strings.TrimSpace(payload.StepID), Need: strings.TrimSpace(repairBreaks(payload.Need))}
}

// A model writing a multi-line `need` sometimes escapes the break twice: the
// value then carries a backslash and an `n` where the break belongs, and the
// notice draws the two characters instead of the break. No line of prose has
// another reading for that pair, so it is repaired where the need is read
// rather than in each frontend that has to draw it.
func repairBreaks(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	// Longest first: a `\r\n` must not be read as its `\r` half and a lone `n`.
	return strings.NewReplacer(`\r\n`, "\n", `\n`, "\n").Replace(s)
}
