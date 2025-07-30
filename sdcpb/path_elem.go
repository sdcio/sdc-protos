package schema_server

import (
	"iter"
	"sort"
)

func NewPathElem(name string, keys map[string]string) *PathElem {
	return &PathElem{
		Name: name,
		Key:  keys,
	}
}

// PathElemNames returns the name of the pathelem plus the values of the alphabetically sorted keys (sorted by key name not key value).
// This will give you the PathElemNames (levels) in the Tree defined by the PathElem.
func (pe *PathElem) PathElemNames() iter.Seq[string] {
	return func(yield func(string) bool) {
		// iterate the patheleme name first
		if !yield(pe.GetName()) {
			return
		}
		pe.PathElemNamesKeysOnly()(yield)
	}
}

func (pe *PathElem) PathElemNamesKeysOnly() iter.Seq[string] {
	return func(yield func(string) bool) {
		keysMap := pe.GetKey()
		keyNameSlice := make([]string, 0, len(keysMap))
		for k := range pe.GetKey() {
			keyNameSlice = append(keyNameSlice, k)
		}
		sort.Strings(keyNameSlice)

		// finally iterate path key values in sorted order
		for _, k := range keyNameSlice {
			if !yield(keysMap[k]) {
				return
			}
		}
	}
}

func (pe *PathElem) DeepCopy() *PathElem {
	result := &PathElem{
		Name: pe.Name,
		Key:  map[string]string{},
	}

	for k, v := range pe.Key {
		result.Key[k] = v
	}

	return result
}
