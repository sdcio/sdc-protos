package tree_persist

import (
	"fmt"
	"strings"

	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
	"google.golang.org/protobuf/proto"
)

func (x *Intent) PrettyString(indent string) string {
	if x == nil {
		return ""
	}
	sb := &strings.Builder{}
	sb.WriteString(fmt.Sprintf("Intent: %s\nPriority: %d\n", x.GetIntentName(), x.GetPriority()))
	x.GetRoot().prettyString(indent, 0, sb)
	return sb.String()
}

func (te *TreeElement) PrettyString(indent string) string {
	if te == nil {
		return ""
	}
	sb := &strings.Builder{}
	te.prettyString(indent, 0, sb)
	return sb.String()
}

func (x *TreeElement) prettyString(indent string, level int, sb *strings.Builder) {
	if x == nil {
		return
	}
	prefix := strings.Repeat(indent, level)
	keylevel := ""
	sb.WriteString(fmt.Sprintf("%s%s%s\n", prefix, x.GetName(), keylevel))
	for _, c := range x.GetChilds() {
		c.prettyString(indent, level+1, sb)
	}
	if len(x.LeafVariant) > 0 {
		tv := &sdcpb.TypedValue{}
		proto.Unmarshal(x.LeafVariant, tv)
		sb.WriteString(prefix + indent + tv.String() + "\n")
	}
}

func (x *TreeElement) CountTerminals() int {
	if x == nil {
		return 0
	}
	counter := 0
	for _, c := range x.Childs {
		counter += c.CountTerminals()
	}
	if x.LeafVariant != nil {
		counter++
	}
	return counter
}
