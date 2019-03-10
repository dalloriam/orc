package keyval

import (
	"encoding/json"
	"errors"
	"fmt"
)

const (
	keyvalModuleName = "keyval"

	keyvalActionGet   = "get"
	keyvalActionSet   = "set"
	keyvalActionClear = "del"
	keyvalActionList  = "list"
)

// Module manages a Key/Value store.
type Module struct {
	keyvalStore map[string]interface{}
}

// NewModule initializes the key/value store.
func NewModule() *Module {
	return &Module{keyvalStore: make(map[string]interface{})}
}

// Name returns the name of the keyval module.
func (m *Module) Name() string { return keyvalModuleName }

// Actions returns the actions supported by the module.
func (m *Module) Actions() []string {
	return []string{keyvalActionGet, keyvalActionSet, keyvalActionClear, keyvalActionList}
}

// Execute executes a key/val action.
func (m *Module) Execute(actionName string, data map[string]interface{}) ([]byte, error) {
	if actionName == keyvalActionList {
		return json.Marshal(map[string]interface{}{"values": m.keyvalStore})
	}

	keyRaw, ok := data["key"]

	if !ok {
		return nil, errors.New("key not specified")
	}

	key := keyRaw.(string)
	val, valOk := data["val"]

	switch actionName {
	case keyvalActionSet:
		if !valOk {
			return nil, errors.New("cannot set, no value specified. use 'val'")
		}
		m.keyvalStore[key] = val
		return json.Marshal(map[string]string{"message": "OK"})

	case keyvalActionGet:
		if returnVal, rOk := m.keyvalStore[key]; rOk {
			return json.Marshal(map[string]interface{}{"value": returnVal})
		}
		return nil, fmt.Errorf("unknown key: %s", key)

	case keyvalActionClear:
		delete(m.keyvalStore, key)
		return json.Marshal(map[string]string{"message": "OK"})
	}

	return nil, fmt.Errorf("unknown action: %s", actionName)
}
