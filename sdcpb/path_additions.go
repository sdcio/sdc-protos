package sdcpb

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
)

var errMalformedXPath = errors.New("malformed xpath")
var errMalformedXPathKey = errors.New("malformed xpath key")

var escapedBracketsReplacer = strings.NewReplacer(`\]`, `]`, `\[`, `[`)

func (p *Path) AddPathElem(pe *PathElem) *Path {
	p.Elem = append(p.Elem, pe)
	return p
}

func (p *Path) CopyPathAddElem(pe *PathElem) *Path {
	child := &Path{
		Origin:      p.Origin,
		Target:      p.Target,
		IsRootBased: p.IsRootBased,
		// Create a new slice header for the child by appending `pe` to the parent's elements.
		// The full-slice expression [:len:len] ensures the new slice has its own header and
		// limited capacity, so appending does not modify the parent's underlying array.
		Elem: append(p.Elem[:len(p.Elem):len(p.Elem)], pe),
	}

	return child
}

func (p *Path) ToXPath(noKeys bool) string {
	if p == nil {
		return ""
	}
	sb := strings.Builder{}
	if p.IsRootBased {
		sb.WriteString("/")
	}
	if p.Origin != "" {
		sb.WriteString(p.Origin)
		sb.WriteString(":")
	}
	elems := p.GetElem()
	numElems := len(elems)
	for i, pe := range elems {
		sb.WriteString(pe.GetName())
		if !noKeys {

			// need to sort the keys to get them in the correct order
			kvMap := pe.GetKey()
			// create a slice for the keys
			keySlice := make([]string, 0, len(pe.GetKey()))
			// add the keys
			for k := range kvMap {
				keySlice = append(keySlice, k)
			}
			// sort the keys
			slices.Sort(keySlice)

			// iterate over the sorted keys slice
			for _, k := range keySlice {
				sb.WriteString("[")
				sb.WriteString(k)
				sb.WriteString("=")
				sb.WriteString(kvMap[k])
				sb.WriteString("]")
			}
		}
		if i+1 != numElems {
			sb.WriteString("/")
		}
	}
	return sb.String()
}

func (p *Path) relativeToAbsPath(currentPath *Path) {
	pElems := make([]*PathElem, 0, len(p.Elem)+len(currentPath.Elem))
	// copy current path to new path
	pElems = append(pElems, currentPath.Elem...)
	for _, pe := range p.GetElem() {
		switch {
		case pe.Name == ".." && len(pElems) != 0:
			// modify new path to follow '..' operations, watching bounds
			pElems = pElems[:len(pElems)-1]
		default:
			// add path element to new path
			pElems = append(pElems, pe)
		}
	}
	// replace the paths Elems with the newly calculated
	p.Elem = pElems
}

func (p *Path) NormalizedAbsPath(currentPath *Path) error {
	if p.hasRelativePathElem() {
		p.relativeToAbsPath(currentPath)
	}
	p.StripPathElemPrefixPath()
	return nil
}

func (p *Path) StripPathElemPrefixPath() {
	for _, pe := range p.GetElem() {
		if i := strings.Index(pe.Name, ":"); i > 0 {
			pe.Name = pe.Name[i+1:]
		}
		// process keys
		for k, v := range pe.Key {
			// delete prefix from key name
			if i := strings.Index(k, ":"); i > 0 {
				delete(pe.Key, k)
				k = k[i+1:]
			}
			// delete prefix from key value
			if strings.Contains(v, ":") {
				kelems := strings.Split(v, "/")
				for idx, kelem := range kelems {
					if i := strings.Index(kelem, ":"); i > 0 {
						kelems[idx] = kelem[i+1:]
					}
				}
				v = strings.Join(kelems, "/")
			}
			pe.Key[k] = v
		}
	}
}

func (p *Path) AbsToRelativePath(refPath *Path) (*Path, error) {
	if refPath == nil || p == nil {
		return nil, fmt.Errorf("AbsToRelativePath: both paths need to be non-nil")
	}
	if !refPath.IsRootBased || !p.IsRootBased {
		return nil, fmt.Errorf("AbsToRelativePath: both paths need to be absolute paths")
	}

	// Find the longest common prefix of elems.
	prefix := commonPrefixLen(p.GetElem(), refPath.GetElem())

	// Number of ".." needed to reach the common ancestor from base.
	up := len(refPath.Elem) - prefix

	var relElems []*PathElem
	relElems = make([]*PathElem, 0, up+(len(p.Elem)-prefix))

	// Up-steps: represented as PathElem{Name: ".."} in gNMI-style relative paths.
	for i := 0; i < up; i++ {
		relElems = append(relElems, &PathElem{Name: ".."})
	}

	// Down-steps: copy the remaining elems from the target.
	for i := prefix; i < len(p.Elem); i++ {
		relElems = append(relElems, p.Elem[i].DeepCopy())
	}

	// If identical paths, return "."
	if len(relElems) == 0 {
		relElems = []*PathElem{}
	}

	return &Path{
		Elem:        relElems,
		IsRootBased: false,
	}, nil
}

// commonPrefixLen takes to PathElem Slices and returns the number of common elements from the root
func commonPrefixLen(a, b []*PathElem) int {
	n := min(len(a), len(b))
	for i := 0; i < n; i++ {
		if !a[i].Equal(b[i]) {
			return i
		}
	}
	return n
}

func (p *Path) LastPathElem() *PathElem {
	return p.GetElem()[len(p.GetElem())-1]
}

func (p *Path) SetIsRootBased(b bool) *Path {
	p.IsRootBased = b
	return p
}

func (p *Path) CopyAndRemoveFirstPathElem() *Path {
	pNew := p.DeepCopy()
	pNew.Elem = pNew.GetElem()[1:]
	return pNew
}

func (p1 *Path) PathsEqual(p2 *Path) bool {
	if p1 == nil && p2 == nil {
		return true
	}
	if p1 == nil || p2 == nil {
		return false
	}
	if len(p1.GetElem()) != len(p2.GetElem()) {
		return false
	}
	for i, pe := range p1.GetElem() {
		if !pe.Equal(p2.GetElem()[i]) {
			return false
		}
	}
	return true
}

func (p *Path) DeepCopy() *Path {
	if p == nil {
		return nil
	}

	result := &Path{
		Origin:      p.Origin,
		Target:      p.Target,
		Elem:        make([]*PathElem, 0, len(p.Elem)),
		IsRootBased: p.IsRootBased,
	}
	// copy each path element
	for _, x := range p.Elem {
		result.Elem = append(result.Elem,
			NewPathElem(x.GetName(), copyMap(x.GetKey())),
		)
	}

	return result
}

func (p *Path) RelativeToAbsPath(currentPath *Path) *Path {
	np := &Path{
		Elem:   make([]*PathElem, 0, len(p.GetElem())+len(currentPath.Elem)),
		Origin: currentPath.GetOrigin(),
		Target: currentPath.GetTarget(),
	}

	// copy current path to new path
	np.Elem = append(np.Elem, currentPath.GetElem()...)

	for _, pe := range p.GetElem() {
		switch {
		case pe.Name == ".." && len(np.Elem) != 0:
			// modify new path to follow '..' operations, watching bounds
			np.Elem = np.Elem[:len(np.Elem)-1]
		default:
			// add path element to new path
			np.Elem = append(np.Elem, pe)
		}
	}

	return np
}

func copyMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	nm := make(map[string]string, len(m))
	for k, v := range m {
		nm[k] = v
	}
	return nm
}

// ParsePath creates a sdcpb.Path out of a p string, check if the first element is prefixed by an origin,
// removes it from the xpath and adds it to the returned sdcpb.Path
func ParsePath(p string) (*Path, error) {
	lp := len(p)
	if lp == 0 {
		return &Path{}, nil
	}
	var origin string
	isRootBased := p[0] == '/'

	idx := strings.Index(p, ":")
	if idx >= 0 && p[0] != '/' && !strings.Contains(p[:idx], "/") &&
		// path == origin:/ || path == origin:
		((idx+1 < lp && p[idx+1] == '/') || (lp == idx+1)) {
		origin = p[:idx]
		p = p[idx+1:]
	}

	pes, err := toPathElems(p)
	if err != nil {
		return nil, err
	}
	return &Path{
		Origin:      origin,
		Elem:        pes,
		IsRootBased: isRootBased,
	}, nil
}

func (p *Path) hasRelativePathElem() bool {
	for _, pe := range p.GetElem() {
		if pe.GetName() == ".." {
			return true
		}
	}
	return false
}

// toPathElems parses a xpath and returns a list of path elements
func toPathElems(p string) ([]*PathElem, error) {
	if !strings.HasSuffix(p, "/") {
		p += "/"
	}
	buffer := make([]rune, 0)
	null := rune(0)
	prevC := rune(0)
	// track if the loop is traversing a key
	inKey := false
	for _, r := range p {
		switch r {
		case '[':
			if inKey && prevC != '\\' {
				return nil, errMalformedXPath
			}
			if prevC != '\\' {
				inKey = true
			}
		case ']':
			if !inKey && prevC != '\\' {
				return nil, errMalformedXPath
			}
			if prevC != '\\' {
				inKey = false
			}
		case '/':
			if !inKey {
				buffer = append(buffer, null)
				prevC = r
				continue
			}
		}
		buffer = append(buffer, r)
		prevC = r
	}
	if inKey {
		return nil, errMalformedXPath
	}
	stringElems := strings.Split(string(buffer), string(null))
	pElems := make([]*PathElem, 0, len(stringElems))
	for _, s := range stringElems {
		if s == "" {
			continue
		}
		pe, err := toPathElem(s)
		if err != nil {
			return nil, err
		}
		pElems = append(pElems, pe)
	}
	return pElems, nil
}

// toPathElem take a xpath formatted path element such as "elem1[k=v]" and returns the corresponding sdcpb.PathElem
func toPathElem(s string) (*PathElem, error) {
	idx := -1
	prevC := rune(0)
	for i, r := range s {
		if r == '[' && prevC != '\\' {
			idx = i
			break
		}
		prevC = r
	}
	var kvs map[string]string
	if idx > 0 {
		var err error
		kvs, err = parseXPathKeys(s[idx:])
		if err != nil {
			return nil, err
		}
		s = s[:idx]
	}
	return &PathElem{Name: s, Key: kvs}, nil
}

// parseXPathKeys takes keys definition from an xpath, e.g [k1=v1][k2=v2] and return the keys and values as a map[string]string
func parseXPathKeys(s string) (map[string]string, error) {
	if len(s) == 0 {
		return nil, nil
	}
	kvs := make(map[string]string)
	inKey := false
	start := 0
	prevRune := rune(0)
	for i, r := range s {
		switch r {
		case '[':
			if prevRune == '\\' {
				prevRune = r
				continue
			}
			if inKey {
				return nil, errMalformedXPathKey
			}
			inKey = true
			start = i + 1
		case ']':
			if prevRune == '\\' {
				prevRune = r
				continue
			}
			if !inKey {
				return nil, errMalformedXPathKey
			}
			eq := strings.Index(s[start:i], "=")
			if eq < 0 {
				return nil, errMalformedXPathKey
			}
			k, v := s[start:i][:eq], s[start:i][eq+1:]
			if len(k) == 0 || len(v) == 0 {
				return nil, errMalformedXPathKey
			}
			kvs[strings.TrimSpace(escapedBracketsReplacer.Replace(k))] = strings.TrimSpace(escapedBracketsReplacer.Replace(v))
			inKey = false
		}
		prevRune = r
	}
	if inKey {
		return nil, errMalformedXPathKey
	}
	return kvs, nil
}

func ComparePath(a, b *Path) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return -1
	}
	if b == nil {
		return 1
	}

	// Compare Origin
	if a.Origin < b.Origin {
		return -1
	}
	if a.Origin > b.Origin {
		return 1
	}

	// Lexicographic comparison of Elem
	minLen := len(a.Elem)
	if len(b.Elem) < minLen {
		minLen = len(b.Elem)
	}
	for i := 0; i < minLen; i++ {
		if cmp := ComparePathElem(a.Elem[i], b.Elem[i]); cmp != 0 {
			return cmp
		}
	}
	// If all common elements are equal, shorter slice wins
	if len(a.Elem) < len(b.Elem) {
		return -1
	}
	if len(a.Elem) > len(b.Elem) {
		return 1
	}

	// Compare Target
	if a.Target < b.Target {
		return -1
	}
	if a.Target > b.Target {
		return 1
	}

	// Compare IsRootBased (false < true)
	if !a.IsRootBased && b.IsRootBased {
		return -1
	}
	if a.IsRootBased && !b.IsRootBased {
		return 1
	}

	// All equal
	return 0
}

// ToStrings converts gnmi.Path to index strings. When index strings are generated,
// gnmi.Path will be irreversibly lost. Index strings will be built by using name field
// in gnmi.PathElem. If gnmi.PathElem has key field, values will be included in
// alphabetical order of the keys.
// E.g. <target>/<origin>/a/b[b:d, a:c]/e will be returned as <target>/<origin>/a/b/c/d/e
// If prefix parameter is set to true, <target> and <origin> fields of
// the gnmi.Path will be prepended in the index strings unless they are empty string.
func ToStrings(p *Path, prefix, nokeys bool) []string {
	is := []string{}
	if p == nil {
		return is
	}
	if prefix {
		// add target to the list of index strings
		if t := p.GetTarget(); t != "" {
			is = append(is, t)
		}
		// add origin to the list of index strings
		if o := p.GetOrigin(); o != "" {
			is = append(is, o)
		}
	}
	for _, e := range p.GetElem() {
		is = append(is, e.GetName())
		if !nokeys {
			is = append(is, sortedVals(e.GetKey())...)
		}
	}

	return is
}

func sortedVals(m map[string]string) []string {
	// Special case single key lists.
	if len(m) == 1 {
		for _, v := range m {
			return []string{v}
		}
	}
	// Return deterministic ordering of multi-key lists.
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	vs := make([]string, 0, len(m))
	for _, k := range ks {
		vs = append(vs, m[k])
	}
	return vs
}
