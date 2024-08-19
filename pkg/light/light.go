package light

import (
	"encoding/json"

	"github.com/utiiz/go-hue/internal/types"
)

type Light struct {
	ID     string `json:"id"`
	State  State  `json:"state"`
	Bridge types.Bridge
}

type State struct {
	On  bool `json:"on"`
	Bri int  `json:"bri"`
	Hue int  `json:"hue"`
	Sat int  `json:"sat"`
}

func NewLight(id string, state State, bridge types.Bridge) *Light {
	return &Light{
		ID:     id,
		State:  state,
		Bridge: bridge,
	}
}

func (l *Light) String() string {
	return l.ID
}

func (l *Light) UnmarshalJSON(data []byte) error {
	var rawMap map[string]json.RawMessage
	err := json.Unmarshal(data, &rawMap)
	if err != nil {
		return err
	}

	var state State
	if stateRaw, ok := rawMap["state"]; ok {
		err = json.Unmarshal(stateRaw, &state)
		if err != nil {
			return err
		}
	}

	l.State = state

	return nil
}

func (l *Light) On() error {
	bridge := l.Bridge

	err := bridge.SetLightOn(l.ID)
	if err != nil {
		return err
	}
	return nil
}
