package schema_server

import (
	"slices"
	"testing"
)

func TestPathElem_PathElemNames(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		peName string
		keys   map[string]string
		want   []string
	}{
		{
			name: "correct order already",
			keys: map[string]string{
				"a": "aval",
				"b": "bval",
				"c": "cval",
				"d": "dval",
				"e": "eval",
				"f": "fval",
				"g": "gval",
			},
			peName: "PE",
			want:   []string{"PE", "aval", "bval", "cval", "dval", "eval", "fval", "gval"},
		},
		{
			name: "inorrect order",
			keys: map[string]string{
				"g": "gval",
				"f": "fval",
				"e": "eval",
				"d": "dval",
				"c": "cval",
				"b": "bval",
				"a": "aval",
			},
			peName: "PE",
			want:   []string{"PE", "aval", "bval", "cval", "dval", "eval", "fval", "gval"},
		},
		{
			// check we are not sorting for values
			name: "inorrect order, reverse val order",
			keys: map[string]string{
				"g": "1val",
				"f": "2val",
				"e": "3val",
				"d": "4val",
				"c": "5val",
				"b": "6val",
				"a": "7val",
			},
			peName: "PE",
			want:   []string{"PE", "7val", "6val", "5val", "4val", "3val", "2val", "1val"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pe := NewPathElem(tt.peName, tt.keys)
			got := pe.PathElemNames()
			result := []string{}
			for x := range got {
				result = append(result, x)
			}

			if slices.Compare(tt.want, result) != 0 {
				t.Errorf("PathElemNames() = %v, want %v", result, tt.want)
			}
		})
	}
}
