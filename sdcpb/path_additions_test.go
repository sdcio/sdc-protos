package sdcpb

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

func TestPath_CopyPathAddKey(t *testing.T) {
	tests := []struct {
		name     string
		path     *Path
		keyName  string
		keyValue string
		want     *Path
	}{
		{
			name: "Add key to last element with no existing keys",
			path: &Path{
				Origin: "origin1",
				Target: "target1",
				Elem: []*PathElem{
					{
						Name: "interface",
						Key: map[string]string{
							"name": "ethernet-1/1",
						},
					},
					{
						Name: "subinterface",
					},
				},
				IsRootBased: true,
			},
			keyName:  "index",
			keyValue: "0",
			want: &Path{
				Origin: "origin1",
				Target: "target1",
				Elem: []*PathElem{
					{
						Name: "interface",
						Key: map[string]string{
							"name": "ethernet-1/1",
						},
					},
					{
						Name: "subinterface",
						Key: map[string]string{
							"index": "0",
						},
					},
				},
				IsRootBased: true,
			},
		},
		{
			name: "Add key to last element with existing keys",
			path: &Path{
				Origin: "origin2",
				Target: "target2",
				Elem: []*PathElem{
					{
						Name: "network-instance",
						Key: map[string]string{
							"name": "default",
						},
					},
					{
						Name: "protocol",
						Key: map[string]string{
							"identifier": "bgp",
						},
					},
				},
				IsRootBased: false,
			},
			keyName:  "name",
			keyValue: "bgp1",
			want: &Path{
				Origin: "origin2",
				Target: "target2",
				Elem: []*PathElem{
					{
						Name: "network-instance",
						Key: map[string]string{
							"name": "default",
						},
					},
					{
						Name: "protocol",
						Key: map[string]string{
							"identifier": "bgp",
							"name":       "bgp1",
						},
					},
				},
				IsRootBased: false,
			},
		},
		{
			name: "Single element path",
			path: &Path{
				Elem: []*PathElem{
					{
						Name: "system",
					},
				},
				IsRootBased: true,
			},
			keyName:  "hostname",
			keyValue: "router1",
			want: &Path{
				Elem: []*PathElem{
					{
						Name: "system",
						Key: map[string]string{
							"hostname": "router1",
						},
					},
				},
				IsRootBased: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Store original path state for comparison
			origXPath := tt.path.ToXPath(false)

			// Call CopyPathAddKey
			got := tt.path.CopyPathAddKey(tt.keyName, tt.keyValue)

			// Verify result matches expected
			if got.ToXPath(false) != tt.want.ToXPath(false) {
				t.Errorf("Path.CopyPathAddKey() = %v, want %v", got.ToXPath(false), tt.want.ToXPath(false))
			}

			// Verify original path was not modified
			if tt.path.ToXPath(false) != origXPath {
				t.Errorf("Path.CopyPathAddKey() modified original path: got %v, want %v", tt.path.ToXPath(false), origXPath)
			}

			// Verify Origin, Target, and IsRootBased are preserved
			if got.Origin != tt.path.Origin {
				t.Errorf("Path.CopyPathAddKey() Origin = %v, want %v", got.Origin, tt.path.Origin)
			}
			if got.Target != tt.path.Target {
				t.Errorf("Path.CopyPathAddKey() Target = %v, want %v", got.Target, tt.path.Target)
			}
			if got.IsRootBased != tt.path.IsRootBased {
				t.Errorf("Path.CopyPathAddKey() IsRootBased = %v, want %v", got.IsRootBased, tt.path.IsRootBased)
			}
		})
	}
}
