package schema_server

import (
	"testing"
)

func TestBlameTreeElement_ToString(t *testing.T) {
	tests := []struct {
		name string
		bte  *BlameTreeElement
		want string
	}{
		{
			name: "single element no value",
			bte: &BlameTreeElement{
				Name: "Root",
			},
		},
		{
			name: "single element with value",
			bte: &BlameTreeElement{
				Name:  "Root",
				Owner: "owner1",
				Value: &TypedValue{Value: &TypedValue_IntVal{IntVal: 6}},
			},
		},
		{
			name: "multiple element",
			bte: &BlameTreeElement{
				Name: "Root",
				Childs: []*BlameTreeElement{
					{
						Name: "interface",
						Childs: []*BlameTreeElement{
							{
								Name: "ethernet-1/2",
								Childs: []*BlameTreeElement{
									{
										Name:  "name",
										Value: &TypedValue{Value: &TypedValue_StringVal{StringVal: "ethernet-1/2"}},
										Owner: "owner2",
									},
									{
										Name:  "description",
										Value: &TypedValue{Value: &TypedValue_StringVal{StringVal: "ethernet-1/2 description"}},
										Owner: "owner2",
									},
								},
							},
							{
								Name: "ethernet-1/1",
								Childs: []*BlameTreeElement{
									{
										Name:  "name",
										Value: &TypedValue{Value: &TypedValue_StringVal{StringVal: "ethernet-1/1"}},
										Owner: "owner1",
									},
									{
										Name:  "admin-state",
										Value: &TypedValue{Value: &TypedValue_StringVal{StringVal: "enable"}},
										Owner: "owner1",
									},
									{
										Name:  "description",
										Value: &TypedValue{Value: &TypedValue_StringVal{StringVal: "ethernet-1/1 description"}},
										Owner: "owner1",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Error(tt.bte.ToString())
			// if got := tt.bte.ToString(); got != tt.want {
			// 	t.Errorf("BlameTreeElement.ToString() = %v, want %v", got, tt.want)
			// }
		})
	}
}
