package sdcpb

import (
	"fmt"
	"iter"
	"os"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

const (
	defaultValueWrapWidth = 96
	inlineValueMaxLen     = 80
	continuationOwnerID   = "..."
	blockValueIndent      = "  "
	minValueWrapWidth     = 24
)

var OriginalLineBreakMarker = "\\n"

type renderedValueLine struct {
	text               string
	hasOriginalNewline bool
}

func normalizeLineBreaks(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}

func shouldRenderValueBlock(s string) bool {
	s = normalizeLineBreaks(s)
	if strings.Contains(s, "\n") {
		return true
	}
	return len([]rune(s)) > inlineValueMaxLen
}

func wrapTextFixedWidth(s string, width int) []string {
	if width <= 0 {
		return []string{s}
	}

	runes := []rune(s)
	if len(runes) <= width {
		return []string{s}
	}

	parts := make([]string, 0, (len(runes)/width)+1)
	for len(runes) > width {
		parts = append(parts, string(runes[:width]))
		runes = runes[width:]
	}
	parts = append(parts, string(runes))
	return parts
}

func splitValueLines(s string, width int) []renderedValueLine {
	s = strings.TrimRight(normalizeLineBreaks(s), "\n")
	if s == "" {
		return []renderedValueLine{{text: ""}}
	}

	raw := strings.Split(s, "\n")
	out := make([]renderedValueLine, 0, len(raw))
	for i, line := range raw {
		wrapped := wrapTextFixedWidth(line, width)
		for j, w := range wrapped {
			out = append(out, renderedValueLine{
				text:               w,
				hasOriginalNewline: i > 0 && j == 0,
			})
		}
	}
	return out
}

func resolveRenderLineWidth() int {
	if v := os.Getenv("SDC_BLAME_WRAP_WIDTH"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}

	if n := detectTerminalWidth(); n > 0 {
		return n
	}

	if v := os.Getenv("COLUMNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return defaultValueWrapWidth
}

func detectTerminalWidth() int {
	for _, fd := range []uintptr{os.Stdout.Fd(), os.Stderr.Fd(), os.Stdin.Fd()} {
		ws, err := unix.IoctlGetWinsize(int(fd), unix.TIOCGWINSZ)
		if err == nil && ws != nil && ws.Col > 0 {
			return int(ws.Col)
		}
	}
	return 0
}

func effectiveValueWrapWidth(ownerSize int, prefix string, isLast bool) int {
	totalLineWidth := resolveRenderLineWidth()
	linePrefix := fmt.Sprintf("%*s%s │ %s%s%s", ownerSize, continuationOwnerID, "   ", prefix, continuationBranch(isLast), blockValueIndent)
	wrapWidth := totalLineWidth - len([]rune(linePrefix)) - len([]rune(OriginalLineBreakMarker))
	if wrapWidth < minValueWrapWidth {
		return minValueWrapWidth
	}
	return wrapWidth
}

func continuationBranch(isLast bool) string {
	if isLast {
		return "    "
	}
	return "│   "
}

func writeValueContinuationLines(sb *strings.Builder, ownerSize int, prefix string, isLast bool, lines []renderedValueLine) {
	branch := continuationBranch(isLast)
	linePrefix := fmt.Sprintf("%*s%s │ %s%s%s", ownerSize, continuationOwnerID, "   ", prefix, branch, blockValueIndent)
	for i, line := range lines {
		sb.WriteString(linePrefix)
		sb.WriteString(line.text)
		if i+1 < len(lines) && lines[i+1].hasOriginalNewline {
			sb.WriteString(OriginalLineBreakMarker)
		}
		sb.WriteString("\n")
	}
}

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
	if len(continuationOwnerID) > maxLen {
		maxLen = len(continuationOwnerID)
	}
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
	deviated := "   "

	switch {
	case b.GetKeyName() != "":
		icon = fmt.Sprintf("🔑 %s=", b.GetKeyName())
	case b.IsDeviated():
		deviated = "(*)"
	case b.GetValue() != nil:
		icon = "🍃 "
	}

	linePrefix := fmt.Sprintf("%*s%s │ %s%s%s%s", ownerSize, b.OwnerNormalized(), deviated, prefix, connector, icon, b.Name)
	// Write this node and value payload.
	if b.IsDeviated() {
		newValue := b.GetDeviationValue().ToString()
		oldValue := b.GetValue().ToString()
		wrapWidth := effectiveValueWrapWidth(ownerSize, prefix, isLast)
		if shouldRenderValueBlock(newValue) || shouldRenderValueBlock(oldValue) {
			sb.WriteString(linePrefix)
			sb.WriteString(" ->")
			sb.WriteString("\n")
			writeValueContinuationLines(sb, ownerSize, prefix, isLast, splitValueLines(newValue, wrapWidth))
			sb.WriteString(fmt.Sprintf("%*s%s │ %s%s%s[~>]\n", ownerSize, continuationOwnerID, "   ", prefix, continuationBranch(isLast), blockValueIndent))
			writeValueContinuationLines(sb, ownerSize, prefix, isLast, splitValueLines(oldValue, wrapWidth))
		} else {
			sb.WriteString(linePrefix)
			sb.WriteString(fmt.Sprintf(" -> %s [~> %s]\n", newValue, oldValue))
		}
	} else if b.GetValue() != nil {
		value := b.GetValue().ToString()
		wrapWidth := effectiveValueWrapWidth(ownerSize, prefix, isLast)
		if shouldRenderValueBlock(value) {
			sb.WriteString(linePrefix)
			sb.WriteString(" ->")
			sb.WriteString("\n")
			writeValueContinuationLines(sb, ownerSize, prefix, isLast, splitValueLines(value, wrapWidth))
		} else {
			sb.WriteString(linePrefix)
			sb.WriteString(" -> ")
			sb.WriteString(value)
			sb.WriteString("\n")
		}
	} else {
		sb.WriteString(linePrefix)
		sb.WriteString("\n")
	}

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
