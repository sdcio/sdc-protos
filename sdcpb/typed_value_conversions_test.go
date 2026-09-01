package sdcpb

import (
	"reflect"
	"testing"

	"google.golang.org/protobuf/proto"
)

func TestXMLRegexConvert(t *testing.T) {

	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "anchors become literals",
			in:   `^\d+$`,
			want: `\^\d+\$`,
		},
		{
			name: "already-escaped anchors stay escaped",
			in:   `foo\$bar`,
			want: `foo\$bar`,
		},
		{
			name: "caret in char class is left alone, dollar is escaped",
			in:   `[^\w]+$`,
			want: `[^\w]+\$`,
		},
		{
			name: "caret later inside char class is escaped",
			in:   `[a^b]`,
			want: `[a\^b]`,
		},
		{
			name: "caret escaped inside char class is escaped",
			in:   `[\^]`,
			want: `[\^]`,
		},
		{
			name: "caret in char class multiple times, dollar is escaped",
			in:   `[^a^b]`,
			want: `[^a\^b]`,
		},
		{
			name: "anchors preceded by a single back-slash stay escaped",
			in:   `\^test\$`,
			want: `\^test\$`,
		},
		{
			name: "empty string",
			in:   ``,
			want: ``,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := xmlRegexConvert(tt.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("XMLRegexConvert() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateBitString(t *testing.T) {
	ref := []*Bit{
		{Name: "a", Position: 0},
		{Name: "b", Position: 0},
		{Name: "c", Position: 0},
	}

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"empty string", "", true},
		{"exact match", "a b c", true},
		{"skipped middle", "a c", true},
		{"single first element", "a", true},
		{"single last element", "c", true},
		{"out of order", "a c b", false},
		{"wrong order overall", "b c a", false},
		{"unknown single element", "d", false},
		{"unknown element in valid", "a c d", false},
		{"duplicate token", "a a", false},
		{"leading / trailing spaces", "  a   b  ", true},
		{"input longer than schema", "a b c d", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := validateBitString(tc.input, ref)
			if got != tc.want {
				t.Fatalf("validateBitString(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestConvertBits(t *testing.T) {
	slt := &SchemaLeafType{
		Bits: []*Bit{{Name: "a", Position: 0}, {Name: "b", Position: 1}, {Name: "c", Position: 2}},
	}
	sltEmpty := &SchemaLeafType{}

	sTv := func(s string) *TypedValue {
		return &TypedValue{Value: &TypedValue_StringVal{StringVal: s}}
	}

	type inStruct struct {
		value string
		slt   *SchemaLeafType
	}

	tests := []struct {
		name    string
		input   inStruct
		want    *TypedValue
		wantErr bool
	}{
		{
			"valid value",
			inStruct{"a b c", slt},
			sTv("a b c"),
			false,
		},
		{
			"invalid value",
			inStruct{"c b a", slt},
			nil,
			true,
		},
		{"empty schema, empty input",
			inStruct{"", sltEmpty},
			nil,
			true,
		},
		{"empty schema, non-empty input",
			inStruct{"a", sltEmpty},
			nil,
			true,
		},
		{"nil schema pointer",
			inStruct{"a", nil},
			nil,
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ConvertBits(tc.input.value, tc.input.slt)
			if tc.wantErr && err == nil {
				t.Fatalf("wanted error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("wanted no error, got %v", err)
			}
			if !proto.Equal(got, tc.want) {
				t.Fatalf("ConvertBits(%q, %q) = %v, want %v", tc.input.value, tc.input.slt, got, tc.want)
			}
		})
	}
}

func TestTVFromStringWithType(t *testing.T) {
	stringSlt := &SchemaLeafType{Type: "string"}
	uint32Slt := &SchemaLeafType{Type: "uint32"}
	boolSlt := &SchemaLeafType{Type: "boolean"}
	leafrefSlt := &SchemaLeafType{
		Type:             "leafref",
		LeafrefTargetType: &SchemaLeafType{Type: "string"},
	}
	unionStringUint32 := &SchemaLeafType{
		Type:       "union",
		UnionTypes: []*SchemaLeafType{stringSlt, uint32Slt},
	}
	nestedUnion := &SchemaLeafType{
		Type: "union",
		UnionTypes: []*SchemaLeafType{
			{
				Type:       "union",
				UnionTypes: []*SchemaLeafType{stringSlt, uint32Slt},
			},
			boolSlt,
		},
	}
	unionWithLeafref := &SchemaLeafType{
		Type:       "union",
		UnionTypes: []*SchemaLeafType{leafrefSlt, uint32Slt},
	}

	tests := []struct {
		name            string
		schemaType      *SchemaLeafType
		value           string
		wantMatchedType *SchemaLeafType
		wantErr         bool
	}{
		{
			name:            "non-union string returns input schemaType",
			schemaType:      stringSlt,
			value:           "hello",
			wantMatchedType: stringSlt,
		},
		{
			name:            "union string-uint32, string value matches string branch",
			schemaType:      unionStringUint32,
			value:           "hello",
			wantMatchedType: stringSlt,
		},
		{
			name:            "union string-uint32, numeric value matches uint32 branch",
			schemaType:      unionStringUint32,
			value:           "42",
			wantMatchedType: stringSlt, // string matches first, so string wins
		},
		{
			name:            "union uint32-string, numeric value matches uint32 first",
			schemaType:      &SchemaLeafType{Type: "union", UnionTypes: []*SchemaLeafType{uint32Slt, stringSlt}},
			value:           "42",
			wantMatchedType: uint32Slt,
		},
		{
			name:       "union with no matching branch returns error",
			schemaType: &SchemaLeafType{Type: "union", UnionTypes: []*SchemaLeafType{uint32Slt, boolSlt}},
			value:      "not-a-number-or-bool",
			wantErr:    true,
		},
		{
			name:            "nested union returns leaf branch (string), not intermediate union",
			schemaType:      nestedUnion,
			value:           "hello",
			wantMatchedType: stringSlt,
		},
		{
			name:            "nested union returns leaf branch (bool) when only bool matches",
			schemaType:      &SchemaLeafType{Type: "union", UnionTypes: []*SchemaLeafType{{Type: "union", UnionTypes: []*SchemaLeafType{uint32Slt}}, boolSlt}},
			value:           "true",
			wantMatchedType: boolSlt,
		},
		{
			name:            "union with leafref branch: matched type is the leafref SchemaLeafType",
			schemaType:      unionWithLeafref,
			value:           "some-string",
			wantMatchedType: leafrefSlt,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tv, matched, err := TVFromStringWithType(tc.schemaType, tc.value, 0)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("wanted error, got nil (tv=%v, matched=%v)", tv, matched)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tv == nil {
				t.Fatalf("expected non-nil TypedValue")
			}
			if matched != tc.wantMatchedType {
				t.Fatalf("matched type = %v, want %v", matched, tc.wantMatchedType)
			}
		})
	}
}

func TestConvertJsonValueToTvWithType(t *testing.T) {
	stringSlt := &SchemaLeafType{Type: "string"}
	uint32Slt := &SchemaLeafType{Type: "uint32"}
	boolSlt := &SchemaLeafType{Type: "boolean"}
	leafrefSlt := &SchemaLeafType{
		Type:             "leafref",
		LeafrefTargetType: &SchemaLeafType{Type: "string"},
	}
	unionStringUint32 := &SchemaLeafType{
		Type:       "union",
		UnionTypes: []*SchemaLeafType{stringSlt, uint32Slt},
	}
	unionUint32String := &SchemaLeafType{
		Type:       "union",
		UnionTypes: []*SchemaLeafType{uint32Slt, stringSlt},
	}
	nestedUnion := &SchemaLeafType{
		Type: "union",
		UnionTypes: []*SchemaLeafType{
			{
				Type:       "union",
				UnionTypes: []*SchemaLeafType{uint32Slt},
			},
			boolSlt,
		},
	}
	unionWithLeafref := &SchemaLeafType{
		Type:       "union",
		UnionTypes: []*SchemaLeafType{leafrefSlt, uint32Slt},
	}

	tests := []struct {
		name            string
		schemaType      *SchemaLeafType
		value           any
		wantMatchedType *SchemaLeafType
		wantErr         bool
	}{
		{
			name:            "non-union string returns input schemaType",
			schemaType:      stringSlt,
			value:           "hello",
			wantMatchedType: stringSlt,
		},
		{
			name:            "union string-uint32, string value matches string branch",
			schemaType:      unionStringUint32,
			value:           "hello",
			wantMatchedType: stringSlt,
		},
		{
			name:            "union uint32-string, numeric value matches uint32 first",
			schemaType:      unionUint32String,
			value:           float64(42),
			wantMatchedType: uint32Slt,
		},
		{
			name:       "union with no matching branch returns error",
			schemaType: &SchemaLeafType{Type: "union", UnionTypes: []*SchemaLeafType{uint32Slt, boolSlt}},
			value:      "not-a-number-or-bool",
			wantErr:    true,
		},
		{
			name:            "nested union returns leaf branch, not intermediate union",
			schemaType:      nestedUnion,
			value:           float64(7),
			wantMatchedType: uint32Slt,
		},
		{
			name:            "union with leafref branch: matched type is the leafref SchemaLeafType",
			schemaType:      unionWithLeafref,
			value:           "some-string",
			wantMatchedType: leafrefSlt,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tv, matched, err := ConvertJsonValueToTvWithType(tc.value, tc.schemaType)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("wanted error, got nil (tv=%v, matched=%v)", tv, matched)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tv == nil {
				t.Fatalf("expected non-nil TypedValue")
			}
			if matched != tc.wantMatchedType {
				t.Fatalf("matched type = %v, want %v", matched, tc.wantMatchedType)
			}
		})
	}
}
