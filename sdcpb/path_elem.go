package sdcpb

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

func (pe1 *PathElem) Equal(pe2 *PathElem) bool {
	return ComparePathElem(pe1, pe2) == 0
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

func ComparePathElem(a, b *PathElem) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return -1
	}
	if b == nil {
		return 1
	}

	// Compare names first
	if a.Name < b.Name {
		return -1
	}
	if a.Name > b.Name {
		return 1
	}

	// Compare maps lexicographically
	akeys := make([]string, 0, len(a.Key))
	for k := range a.Key {
		akeys = append(akeys, k)
	}
	bkeys := make([]string, 0, len(b.Key))
	for k := range b.Key {
		bkeys = append(bkeys, k)
	}
	sort.Strings(akeys)
	sort.Strings(bkeys)

	minLen := len(akeys)
	if len(bkeys) < minLen {
		minLen = len(bkeys)
	}
	for i := 0; i < minLen; i++ {
		if akeys[i] < bkeys[i] {
			return -1
		}
		if akeys[i] > bkeys[i] {
			return 1
		}
		// compare values if keys equal
		if a.Key[akeys[i]] < b.Key[bkeys[i]] {
			return -1
		}
		if a.Key[akeys[i]] > b.Key[bkeys[i]] {
			return 1
		}
	}

	// If all common keys are equal, shorter map wins
	if len(akeys) < len(bkeys) {
		return -1
	}
	if len(akeys) > len(bkeys) {
		return 1
	}

	// All equal
	return 0
}

func (p *PathElem) AddKey(key, value string) {
	if p.Key == nil {
		p.Key = map[string]string{}
	}
	p.Key[key] = value
}
