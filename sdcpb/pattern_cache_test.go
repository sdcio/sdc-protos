package sdcpb

import (
	"strings"
	"sync"
	"testing"
)

func stringType(patterns ...*SchemaPattern) *SchemaLeafType {
	return &SchemaLeafType{Type: "string", Patterns: patterns}
}

func TestCompiledPatternCacheHit(t *testing.T) {
	resetPatternCache()
	re1, err := compiledPattern(`[a-z]+`)
	if err != nil {
		t.Fatal(err)
	}
	re2, err := compiledPattern(`[a-z]+`)
	if err != nil {
		t.Fatal(err)
	}
	if re1 != re2 {
		t.Fatal("expected the same compiled pattern on second lookup")
	}
}

func TestConvertStringPatterns(t *testing.T) {
	resetPatternCache()
	tests := []struct {
		name     string
		value    string
		patterns []*SchemaPattern
		wantErr  bool
	}{
		{"no patterns", "anything", nil, false},
		{"match", "abc", []*SchemaPattern{{Pattern: `[a-z]+`}}, false},
		{"no match", "ABC", []*SchemaPattern{{Pattern: `[a-z]+`}}, true},
		{"inverted, matches so invalid", "abc", []*SchemaPattern{{Pattern: `[a-z]+`, Inverted: true}}, true},
		{"inverted, no match so valid", "ABC", []*SchemaPattern{{Pattern: `[a-z]+`, Inverted: true}}, false},
		{"anded, all match", "abc", []*SchemaPattern{{Pattern: `[a-z]+`}, {Pattern: `a.*`}}, false},
		{"anded, one fails", "abc", []*SchemaPattern{{Pattern: `[a-z]+`}, {Pattern: `x.*`}}, true},
		{"same pattern inverted and not, repeated", "abc", []*SchemaPattern{{Pattern: `[a-z]+`}, {Pattern: `[a-z]+`, Inverted: true}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// run twice: cold and warm cache must agree
			for i := 0; i < 2; i++ {
				_, err := ConvertString(tt.value, stringType(tt.patterns...))
				if (err != nil) != tt.wantErr {
					t.Fatalf("iteration %d: err=%v wantErr=%v", i, err, tt.wantErr)
				}
			}
		})
	}
}

func TestConvertStringInvalidPatternRepeated(t *testing.T) {
	resetPatternCache()
	lst := stringType(&SchemaPattern{Pattern: `(unclosed`})
	var first string
	for i := 0; i < 3; i++ {
		_, err := ConvertString("x", lst)
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.HasPrefix(err.Error(), "unable to compile regex: ") {
			t.Fatalf("unexpected error text: %v", err)
		}
		if i == 0 {
			first = err.Error()
		} else if err.Error() != first {
			t.Fatalf("error changed: %q vs %q", err.Error(), first)
		}
	}
}

func TestConvertStringConcurrent(t *testing.T) {
	resetPatternCache()
	lst := stringType(&SchemaPattern{Pattern: `[a-z]+`}, &SchemaPattern{Pattern: `(bad`})
	ok := stringType(&SchemaPattern{Pattern: `[a-z]+`})
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				if _, err := ConvertString("abc", ok); err != nil {
					t.Error(err)
					return
				}
				if _, err := ConvertString("abc", lst); err == nil {
					t.Error("expected error")
					return
				}
			}
		}()
	}
	wg.Wait()
}
