package sdcpb

import (
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestSchemaElem_IsMutateOnWrite(t *testing.T) {
	tests := []struct {
		name string
		elem *SchemaElem
		want bool
	}{
		{
			name: "nil SchemaElem",
			elem: nil,
			want: false,
		},
		{
			name: "unset SchemaElem oneof",
			elem: &SchemaElem{},
			want: false,
		},
		{
			name: "container returns false",
			elem: &SchemaElem{
				Schema: &SchemaElem_Container{
					Container: &ContainerSchema{Name: "system"},
				},
			},
			want: false,
		},
		{
			name: "leaf with nil Field returns false",
			elem: &SchemaElem{
				Schema: &SchemaElem_Field{},
			},
			want: false,
		},
		{
			name: "leaf with MutateOnWrite false and Encrypted true returns false",
			elem: &SchemaElem{
				Schema: &SchemaElem_Field{
					Field: &LeafSchema{
						Name:          "password",
						Encrypted:     true,
						MutateOnWrite: false,
					},
				},
			},
			want: false,
		},
		{
			name: "leaf with MutateOnWrite true returns true",
			elem: &SchemaElem{
				Schema: &SchemaElem_Field{
					Field: &LeafSchema{
						Name:          "password",
						MutateOnWrite: true,
					},
				},
			},
			want: true,
		},
		{
			name: "leaflist with nil Leaflist returns false",
			elem: &SchemaElem{
				Schema: &SchemaElem_Leaflist{},
			},
			want: false,
		},
		{
			name: "leaflist with MutateOnWrite false returns false",
			elem: &SchemaElem{
				Schema: &SchemaElem_Leaflist{
					Leaflist: &LeafListSchema{
						Name:          "keys",
						Encrypted:     true,
						MutateOnWrite: false,
					},
				},
			},
			want: false,
		},
		{
			name: "leaflist with MutateOnWrite true returns true",
			elem: &SchemaElem{
				Schema: &SchemaElem_Leaflist{
					Leaflist: &LeafListSchema{
						Name:          "keys",
						MutateOnWrite: true,
					},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.elem.IsMutateOnWrite(); got != tt.want {
				t.Fatalf("IsMutateOnWrite() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMutateOnWriteProtoFieldNumbers(t *testing.T) {
	checkMsg := func(t *testing.T, desc protoreflect.MessageDescriptor) {
		t.Helper()
		mow := desc.Fields().ByName("mutate_on_write")
		if mow == nil {
			t.Fatalf("%s missing mutate_on_write field", desc.FullName())
		}
		if mow.Number() != 26 {
			t.Fatalf("%s.mutate_on_write number = %d, want 26", desc.FullName(), mow.Number())
		}
		if mow.Kind() != protoreflect.BoolKind {
			t.Fatalf("%s.mutate_on_write kind = %v, want bool", desc.FullName(), mow.Kind())
		}

		enc := desc.Fields().ByName("encrypted")
		if enc == nil || enc.Number() != 7 {
			t.Fatalf("%s.encrypted field missing or moved from tag 7: %v", desc.FullName(), enc)
		}
	}

	checkMsg(t, (&LeafSchema{}).ProtoReflect().Descriptor())
	checkMsg(t, (&LeafListSchema{}).ProtoReflect().Descriptor())
}
