package tree_persist_test

import (
	"testing"

	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
	"github.com/sdcio/sdc-protos/tree_persist"
	"google.golang.org/protobuf/proto"
)

// TestIntentSensitivePathsRoundTrip verifies that sensitive_paths survives a
// proto marshal/unmarshal cycle, and that decoding an old intent blob (one
// that was serialised without the field) yields an empty slice — no migration
// required.
func TestIntentSensitivePathsRoundTrip(t *testing.T) {
	t.Run("populated sensitive_paths survives marshal/unmarshal", func(t *testing.T) {
		want := []*sdcpb.Path{
			{Elem: []*sdcpb.PathElem{{Name: "bgp"}, {Name: "neighbors"}, {Name: "auth-password"}}, IsRootBased: true},
			{Elem: []*sdcpb.PathElem{{Name: "system"}, {Name: "aaa"}, {Name: "secret"}}, IsRootBased: true},
		}
		orig := &tree_persist.Intent{
			IntentName:     "test-intent",
			Priority:       100,
			SensitivePaths: want,
		}

		b, err := proto.Marshal(orig)
		if err != nil {
			t.Fatalf("proto.Marshal: %v", err)
		}

		got := &tree_persist.Intent{}
		if err := proto.Unmarshal(b, got); err != nil {
			t.Fatalf("proto.Unmarshal: %v", err)
		}

		if len(got.GetSensitivePaths()) != len(want) {
			t.Fatalf("got %d sensitive_paths, want %d", len(got.GetSensitivePaths()), len(want))
		}
		for i, p := range want {
			if got.GetSensitivePaths()[i].ToXPath(true) != p.ToXPath(true) {
				t.Errorf("sensitive_paths[%d]: got %q, want %q", i, got.GetSensitivePaths()[i].ToXPath(true), p.ToXPath(true))
			}
		}
	})

	t.Run("absent sensitive_paths in legacy blob decodes as empty slice", func(t *testing.T) {
		// Serialise an intent without the sensitive_paths field to simulate an
		// existing intent stored before this feature was introduced.
		legacy := &tree_persist.Intent{
			IntentName: "legacy-intent",
			Priority:   50,
		}
		b, err := proto.Marshal(legacy)
		if err != nil {
			t.Fatalf("proto.Marshal: %v", err)
		}

		got := &tree_persist.Intent{}
		if err := proto.Unmarshal(b, got); err != nil {
			t.Fatalf("proto.Unmarshal: %v", err)
		}

		if paths := got.GetSensitivePaths(); len(paths) != 0 {
			t.Errorf("expected empty sensitive_paths for legacy blob, got %v", paths)
		}
	})
}
