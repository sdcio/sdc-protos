package sdcpb

import (
	"fmt"
	"iter"
	"sort"
	"strings"
)

func NewBlameTreeElement(name string) *BlameTreeElement {
	return &BlameTreeElement{Name: name}
}

// Copy creates a shallow copy of the BlameTreeElement, including all its fields but not its children.
// This is useful for creating modified versions of an element without affecting the original.
func (b *BlameTreeElement) Copy() *BlameTreeElement {
	if b == nil {
		return nil
	}
	newElem := &BlameTreeElement{
		Name:           b.Name,
		KeyName:        b.KeyName,
		Value:          b.Value,
		DeviationValue: b.DeviationValue,
		Owner:          b.Owner,
	}
	return newElem
}

func (b *BlameTreeElement) SetValue(tv *TypedValue) *BlameTreeElement {
	b.Value = tv
	return b
}

func (b *BlameTreeElement) SetOwner(owner string) *BlameTreeElement {
	b.Owner = owner
	return b
}

func (b *BlameTreeElement) SetDeviationValue(tv *TypedValue) *BlameTreeElement {
	b.DeviationValue = tv
	return b
}

func (b *BlameTreeElement) SetKeyName(keyName string) *BlameTreeElement {
	b.KeyName = keyName
	return b
}

func (b *BlameTreeElement) GetPath(parentPath *Path) *Path {
	var result *Path = nil
	// If nil is provided, we assume this is the root element and create a new path with "root" as the first element
	if parentPath == nil {
		parentPath = &Path{Elem: []*PathElem{}, IsRootBased: true}
	}
	// Create a copy of the parent path to avoid mutating it
	if b.GetKeyName() == "" {
		result = parentPath.CopyPathAddElem(NewPathElem(b.Name, nil))
	} else {
		result = parentPath.CopyPathAddKey(b.GetKeyName(), b.GetName())
	}

	return result
}

func (b *BlameTreeElement) CalculateMaxOwnerLength() int {
	maxLen := len(b.OwnerNormalized())
	for _, c := range b.GetChilds() {
		childMax := c.CalculateMaxOwnerLength()
		if childMax > maxLen {
			maxLen = childMax
		}
	}
	return maxLen
}

func (b *BlameTreeElement) AddChild(c *BlameTreeElement) *BlameTreeElement {
	b.Childs = append(b.Childs, c)
	return b
}

func (b *BlameTreeElement) StringIndent(sb *strings.Builder, prefix string, isLast bool, isRoot bool, ownerSize int) string {
	// Tree-style connectors
	connector := "├── "
	nextPrefix := prefix + "│   "
	icon := "📦 "

	if isRoot {
		connector = "    "
		nextPrefix = prefix + "    "
		icon = "🎯 "
	}

	if isLast {
		connector = "└── "
		nextPrefix = prefix + "    "
	}

	// Compose value string
	value := ""
	deviated := "   "
	deviated_value := ""

	switch {
	case b.GetKeyName() != "":
		icon = fmt.Sprintf("🔑 %s=", b.GetKeyName())
	case b.IsDeviated():
		deviated = "(*)"
		value = fmt.Sprintf(" -> %s", b.GetDeviationValue().ToString())
		deviated_value = fmt.Sprintf(" [~> %s]", b.GetValue().ToString())
	case b.GetValue() != nil:
		value = fmt.Sprintf(" -> %s", b.GetValue().ToString())
		icon = "🍃 "
	}

	// Write this node
	sb.WriteString(fmt.Sprintf("%*s%s │ %s%s%s%s%s%s\n", ownerSize, b.OwnerNormalized(), deviated, prefix, connector, icon, b.Name, value, deviated_value))

	// Write children
	cCount := b.ChildCount()
	counter := 0
	for c := range b.SortedChildIterator() {
		counter++
		c.StringIndent(sb, nextPrefix, cCount == counter, false, ownerSize)
	}
	return sb.String()
}

func (b *BlameTreeElement) IsDeviated() bool {
	return b.GetDeviationValue() != nil
}

func (b *BlameTreeElement) ToString() string {
	if b == nil {
		return ""
	}
	sb := &strings.Builder{}
	// Root node typically has no prefix or connector
	b.StringIndent(sb, "", false, true, b.CalculateMaxOwnerLength())
	return strings.TrimSuffix(sb.String(), "\n")
}

func (b *BlameTreeElement) OwnerNormalized() string {
	if b.Owner == "" {
		return "-----"
	}
	return b.Owner
}

func (b *BlameTreeElement) SortedChildIterator() iter.Seq[*BlameTreeElement] {
	return func(yield func(*BlameTreeElement) bool) {

		// Sort by Name
		sort.Slice(b.Childs, func(i, j int) bool {
			return b.Childs[i].Name < b.Childs[j].Name
		})

		// Yield each child
		for _, child := range b.Childs {
			if !yield(child) {
				return
			}
		}
	}
}

func (b *BlameTreeElement) ChildCount() int {
	return len(b.Childs)
}

func (b *BlameTreeElement) StringXPath() string {
	if b == nil {
		return ""
	}
	sb := &strings.Builder{}

	maxOwnerLength := b.CalculateMaxOwnerLength()

	// Root node typically has no prefix or connector
	b.WalkPath(&Path{Elem: nil, IsRootBased: true}, func(elem *BlameTreeElement, path *Path) {
		if elem.StringXPathSingle(sb, path, maxOwnerLength) {
			sb.WriteString("\n")
		}
	})
	return strings.TrimSuffix(sb.String(), "\n")
}

func (b *BlameTreeElement) StringSliceXPath() []string {
	if b == nil {
		return []string{}
	}
	maxOwnerLength := b.CalculateMaxOwnerLength()

	result := []string{}
	sb := &strings.Builder{}

	// Root node typically has no prefix or connector
	b.WalkPath(&Path{Elem: nil, IsRootBased: true}, func(elem *BlameTreeElement, path *Path) {
		mustAdd := elem.StringXPathSingle(sb, path, maxOwnerLength)
		if mustAdd {
			result = append(result, sb.String())
		}
		sb.Reset()
	})
	return result
}

func (b *BlameTreeElement) WalkPath(path *Path, fn func(*BlameTreeElement, *Path)) {
	if b == nil {
		return
	}
	fn(b, path)

	for _, c := range b.GetChilds() {
		var childPath *Path
		if c.GetKeyName() != "" {
			childPath = path.CopyPathAddKey(c.GetKeyName(), c.GetName())
		} else {
			childPath = path.CopyPathAddElem(NewPathElem(c.GetName(), nil))
		}
		c.WalkPath(childPath, fn)
	}
}

func (b *BlameTreeElement) StringXPathSingle(sb *strings.Builder, path *Path, maxOwnerLength int) bool {
	val := b.GetValue()
	if val != nil {
		deviationTxt := ""
		deviated := " "
		if b.IsDeviated() {
			deviationTxt = fmt.Sprintf(" [~> %s]", val.ToString())
			val = b.GetDeviationValue()
			deviated = "D"
		}

		fmt.Fprintf(sb, "%s [ %-*s ] %s -> %s%s", deviated, maxOwnerLength, b.GetOwner(), path.ToXPath(false), val.ToString(), deviationTxt)
		return true
	}
	return false
}
