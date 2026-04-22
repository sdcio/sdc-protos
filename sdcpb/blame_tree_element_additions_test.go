package sdcpb

import (
	"strings"
	"testing"
)

func TestToStringMultilineLeafValueKeepsTreePrefix(t *testing.T) {
	root := NewBlameTreeElement("root")
	root.AddChild(
		NewBlameTreeElement("certificate").
			SetOwner("running").
			SetValue(&TypedValue{Value: &TypedValue_StringVal{StringVal: "line1\nline2\nline3"}}),
	)

	out := root.ToString()
	lines := strings.Split(out, "\n")

	if len(lines) < 4 {
		t.Fatalf("unexpected output line count: got %d, output:\n%s", len(lines), out)
	}

	if !strings.Contains(out, "certificate ->") {
		t.Fatalf("expected header line for multiline value, got:\n%s", out)
	}
	if !strings.Contains(out, "...") {
		t.Fatalf("expected continuation owner marker, got:\n%s", out)
	}
	if !strings.Contains(out, "line1\\n") || !strings.Contains(out, "line2\\n") {
		t.Fatalf("expected original newline markers in continuation lines, got:\n%s", out)
	}

	for _, line := range lines {
		if strings.Contains(line, "line2") || strings.Contains(line, "line3") {
			if !strings.Contains(line, "│") {
				t.Fatalf("continuation line lost tree prefix: %q", line)
			}
			if strings.HasPrefix(line, "line2") || strings.HasPrefix(line, "line3") {
				t.Fatalf("continuation line is unprefixed: %q", line)
			}
		}
	}
}

func TestToStringMultilineDeviatedValueKeepsTreePrefix(t *testing.T) {
	root := NewBlameTreeElement("root")
	root.AddChild(
		NewBlameTreeElement("certificate").
			SetOwner("running").
			SetValue(&TypedValue{Value: &TypedValue_StringVal{StringVal: "old1\nold2"}}).
			SetDeviationValue(&TypedValue{Value: &TypedValue_StringVal{StringVal: "new1\nnew2"}}),
	)

	out := root.ToString()
	lines := strings.Split(out, "\n")

	if !strings.Contains(out, "certificate ->") {
		t.Fatalf("expected deviated header line, got:\n%s", out)
	}
	if !strings.Contains(out, "[~>]") {
		t.Fatalf("expected output to contain old value marker, got:\n%s", out)
	}

	for _, line := range lines {
		if strings.Contains(line, "new2") || strings.Contains(line, "old2") {
			if !strings.Contains(line, "│") {
				t.Fatalf("continuation line lost tree prefix: %q", line)
			}
			if strings.HasPrefix(line, "new2") || strings.HasPrefix(line, "old2") {
				t.Fatalf("continuation line is unprefixed: %q", line)
			}
		}
	}
}

func TestToStringLongLeafValueWrapsWithPrefix(t *testing.T) {
	root := NewBlameTreeElement("root")
	longLine := strings.Repeat("A", defaultValueWrapWidth+20)
	root.AddChild(
		NewBlameTreeElement("key").
			SetOwner("running").
			SetValue(&TypedValue{Value: &TypedValue_StringVal{StringVal: longLine}}),
	)

	out := root.ToString()
	lines := strings.Split(out, "\n")

	if len(lines) < 3 {
		t.Fatalf("expected wrapped continuation line, got %d lines:\n%s", len(lines), out)
	}
	if !strings.Contains(out, "key ->") {
		t.Fatalf("expected long value to switch to block mode, got:\n%s", out)
	}

	wrappedChunk := strings.Repeat("A", 20)
	foundWrapped := false
	for _, line := range lines {
		if strings.Contains(line, wrappedChunk) && strings.Contains(line, "│") {
			foundWrapped = true
			break
		}
	}
	if !foundWrapped {
		t.Fatalf("expected wrapped continuation chunk with tree prefix, got:\n%s", out)
	}
	if strings.Contains(out, "\\n ") {
		t.Fatalf("did not expect original newline marker for soft wraps, got:\n%s", out)
	}
}

func TestToStringTrailingNewlineDoesNotEmitExtraEmptyContinuationLine(t *testing.T) {
	root := NewBlameTreeElement("root")
	root.AddChild(
		NewBlameTreeElement("certificate").
			SetOwner("running").
			SetValue(&TypedValue{Value: &TypedValue_StringVal{StringVal: "line1\nline2\n"}}),
	)

	out := root.ToString()
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.TrimSpace(strings.TrimPrefix(line, "...")) == "│" {
			t.Fatalf("found empty continuation line: %q", line)
		}
	}
}

func TestToStringContinuationKeepsSiblingBranchVisible(t *testing.T) {
	root := NewBlameTreeElement("root")
	profile := NewBlameTreeElement("profile")
	profile.AddChild(
		NewBlameTreeElement("certificate").
			SetOwner("running").
			SetValue(&TypedValue{Value: &TypedValue_StringVal{StringVal: "line1\nline2"}}),
	)
	profile.AddChild(
		NewBlameTreeElement("cipher-list").
			SetOwner("default").
			SetValue(&TypedValue{Value: &TypedValue_StringVal{StringVal: "c1"}}),
	)
	root.AddChild(profile)

	out := root.ToString()

	if !strings.Contains(out, "certificate ->") || !strings.Contains(out, "cipher-list -> c1") {
		t.Fatalf("expected both certificate block and sibling line, got:\n%s", out)
	}

	if !strings.Contains(out, "...") || !strings.Contains(out, "line1\\n") || !strings.Contains(out, "│     line2") {
		t.Fatalf("expected continuation line to keep branch pipe, got:\n%s", out)
	}
}
