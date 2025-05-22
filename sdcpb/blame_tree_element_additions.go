package schema_server

import (
	"fmt"
	"iter"
	"sort"
	"strings"
)

func NewBlameTreeElement(name string) *BlameTreeElement {
	return &BlameTreeElement{Name: name}
}

func (b *BlameTreeElement) SetValue(tv *TypedValue) *BlameTreeElement {
	b.Value = tv
	return b
}

func (b *BlameTreeElement) SetOwner(owner string) *BlameTreeElement {
	b.Owner = owner
	return b
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
	if b.GetValue() != nil {
		value = fmt.Sprintf(" -> %s", b.Value.ToString())
		icon = "🍃 "
	}

	// Write this node
	sb.WriteString(fmt.Sprintf("%*s  │  %s%s%s%s%s\n", ownerSize, b.OwnerNormalized(), prefix, connector, icon, b.Name, value))

	// Write children
	cCount := b.ChildCount()
	counter := 0
	for c := range b.SortedChildIterator() {
		counter++
		c.StringIndent(sb, nextPrefix, cCount == counter, false, ownerSize)
	}
	return sb.String()
}

func (b *BlameTreeElement) ToString() string {
	if b == nil {
		return "<nil>"
	}
	sb := &strings.Builder{}
	// Root node typically has no prefix or connector
	b.StringIndent(sb, "", false, true, b.CalculateMaxOwnerLength())
	return sb.String()
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
