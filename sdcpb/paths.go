package sdcpb

type Paths []*Path

// ToXPathSlice converts the Paths slice to a slice of XPath strings.
func (p Paths) ToXPathSlice() []string {
	result := make([]string, 0, len(p))
	for _, x := range p {
		result = append(result, x.ToXPath(false))
	}
	return result
}

// ContainsParentPath checks if any path in the Paths slice is a parent path of the given path.
func (p Paths) ContainsParentPath(path *Path) bool {
	for _, x := range p {
		if x.IsParentPathOf(path) {
			return true
		}
	}
	return false
}

// DeepCopy creates a deep copy of the Paths slice, including deep copies of each Path and its PathElems.
func (p Paths) DeepCopy() Paths {
	result := make(Paths, len(p))
	for i, path := range p {
		result[i] = path.DeepCopy()
	}
	return result
}
