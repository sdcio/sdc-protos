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

func TestConvertBoolean(t *testing.T) {
	bv := func(b bool) *TypedValue {
		return &TypedValue{Value: &TypedValue_BoolVal{BoolVal: b}}
	}
	tests := []struct {
		name    string
		input   string
		want    *TypedValue
		wantErr bool
	}{
		{"lowercase true", "true", bv(true), false},
		{"lowercase false", "false", bv(false), false},
		{"Title True", "True", bv(true), false},
		{"Title False", "False", bv(false), false},
		{"UPPER TRUE", "TRUE", bv(true), false},
		{"UPPER FALSE", "FALSE", bv(false), false},
		{"numeric 1", "1", bv(true), false},
		{"numeric 0", "0", bv(false), false},
		{"leading/trailing whitespace", " true ", bv(true), false},
		{"invalid", "yes", nil, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ConvertBoolean(tc.input, nil)
			if tc.wantErr && err == nil {
				t.Fatalf("wanted error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("wanted no error, got %v", err)
			}
			if !proto.Equal(got, tc.want) {
				t.Fatalf("ConvertBoolean(%q) = %v, want %v", tc.input, got, tc.want)
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
