package config

// Keybindings holds configuration for key bindings.
type Keybindings struct {
	Actions []Action `yaml:"actions"` // List of configurable actions
}

// FindActionByKey finds an action by its key binding.
func (k *Keybindings) FindActionByKey(key string) *Action {
	for i, action := range k.Actions {
		if action.Key == key {
			return &k.Actions[i]
		}
	}
	return nil
}

// GetActionKeys returns all configured action keys for help display.
func (k *Keybindings) GetActionKeys() []string {
	var keys []string
	for _, action := range k.Actions {
		keys = append(keys, action.Key)
	}
	return keys
}
