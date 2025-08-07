package schema_server

import (
	"testing"
)

func TestPath_AbsToRelativePath(t *testing.T) {
	tests := []struct {
		name    string
		path    *Path
		refPath *Path
		want    *Path
		wantErr bool
	}{
		{
			name: "Refpath is Root",
			path: &Path{
				Elem: []*PathElem{
					{
						Name: "interface",
						Key: map[string]string{
							"name": "ethernet-1/1",
						},
					},
					{
						Name: "admin-state",
					},
				},
				IsRootBased: true,
			},
			refPath: &Path{
				Elem:        []*PathElem{},
				IsRootBased: true,
			},
			want: &Path{
				Elem: []*PathElem{
					{
						Name: "interface",
						Key: map[string]string{
							"name": "ethernet-1/1",
						},
					},
					{
						Name: "admin-state",
					},
				},
			},
		},
		{
			name: "Refpath not isRootBased path",
			path: &Path{
				Elem: []*PathElem{
					{
						Name: "interface",
						Key: map[string]string{
							"name": "ethernet-1/1",
						},
					},
					{
						Name: "admin-state",
					},
				},
				IsRootBased: true,
			},
			refPath: &Path{
				Elem: []*PathElem{},
			},
			wantErr: true,
		},
		{
			name: "differet interface",
			path: &Path{
				Elem: []*PathElem{
					{
						Name: "interface",
						Key: map[string]string{
							"name": "ethernet-1/2",
						},
					},
					{
						Name: "admin-state",
					},
				},
				IsRootBased: true,
			},
			refPath: &Path{
				Elem: []*PathElem{
					{
						Name: "interface",
						Key: map[string]string{
							"name": "ethernet-1/1",
						},
					},
					{
						Name: "admin-state",
					},
				},
				IsRootBased: true,
			},
			want: &Path{
				Elem: []*PathElem{
					{
						Name: "..",
					},
					{
						Name: "..",
					},
					{
						Name: "interface",
						Key: map[string]string{
							"name": "ethernet-1/2",
						},
					},
					{
						Name: "admin-state",
					},
				},
			},
		},
		{
			name: "same path",
			path: &Path{
				Elem: []*PathElem{
					{
						Name: "interface",
						Key: map[string]string{
							"name": "ethernet-1/2",
						},
					},
					{
						Name: "admin-state",
					},
				},
				IsRootBased: true,
			},
			refPath: &Path{
				Elem: []*PathElem{
					{
						Name: "interface",
						Key: map[string]string{
							"name": "ethernet-1/2",
						},
					},
					{
						Name: "admin-state",
					},
				},
				IsRootBased: true,
			},
			want: &Path{
				Elem: []*PathElem{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.path.AbsToRelativePath(tt.refPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Path.AbsToRelativePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.ToXPath(false) != tt.want.ToXPath(false) {
				t.Errorf("Path.AbsToRelativePath() = %v, want %v", got.ToXPath(false), tt.want.ToXPath(false))
			}
		})
	}
}
