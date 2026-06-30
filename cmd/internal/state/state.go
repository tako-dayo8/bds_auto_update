package state

import (
	"encoding/json"
	"os"
	"time"
)

const STATE_FILE_PATH = "./state.json"

type State struct {
	Version   string     `json:"version"`
	InstallAt *time.Time `json:"install_at"`
}

func LoadState() (*State, error) {
	data, err := os.ReadFile(STATE_FILE_PATH)
	if err != nil {
		defaultState := State{}
		data, err = json.Marshal(defaultState)
		if err != nil {
			return nil, err
		}
		if err = os.WriteFile(STATE_FILE_PATH, data, 0644); err != nil {
			return nil, err
		}
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

func WriteState(state State) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	if err := os.WriteFile(STATE_FILE_PATH, []byte(data), 0644); err != nil {
		return err
	}

	return nil
}
