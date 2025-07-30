package schema_server

type PathSet struct {
	pathMap map[string]*Path
}

func NewPathSet() *PathSet {
	return &PathSet{
		pathMap: map[string]*Path{},
	}
}

func (ps *PathSet) AddPath(p *Path) {
	key := p.ToXPath(false)
	if _, exists := ps.pathMap[key]; !exists {
		ps.pathMap[key] = p
	}
}

func (ps *PathSet) Join(otherPs *PathSet) {
	for k, v := range otherPs.pathMap {
		ps.pathMap[k] = v
	}
}
