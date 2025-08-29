package schema_server

import "iter"

type PathSet struct {
	pathMap map[string]*Path
}

func NewPathSet() *PathSet {
	return &PathSet{
		pathMap: map[string]*Path{},
	}
}

func (ps *PathSet) DeepCopy() *PathSet {
	result := NewPathSet()
	for k, v := range ps.pathMap {
		result.pathMap[k] = v.DeepCopy()
	}
	return result
}

func (ps *PathSet) AddPaths(paths []*Path) *PathSet {
	for _, p := range paths {
		ps.AddPath(p)
	}
	return ps
}

func (ps *PathSet) AddPath(p *Path) *PathSet {
	key := p.ToXPath(false)
	if _, exists := ps.pathMap[key]; !exists {
		ps.pathMap[key] = p
	}
	return ps
}

func (ps *PathSet) Join(otherPs *PathSet) *PathSet {
	for k, v := range otherPs.pathMap {
		ps.pathMap[k] = v
	}
	return ps
}

func (ps *PathSet) Items() iter.Seq[*Path] {
	return func(yield func(*Path) bool) {
		for _, v := range ps.pathMap {
			if !yield(v) {
				return
			}
		}
	}
}
