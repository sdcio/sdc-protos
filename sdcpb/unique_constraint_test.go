package sdcpb

import (
	"testing"
)

func TestContainerSchema_GetUniqueConstraints(t *testing.T) {
	t.Run("returns nil for container with no unique constraints", func(t *testing.T) {
		cs := &ContainerSchema{}
		if got := cs.GetUniqueConstraints(); got != nil {
			t.Errorf("GetUniqueConstraints() = %v, want nil", got)
		}
	})

	t.Run("single unique constraint with two elements", func(t *testing.T) {
		cs := &ContainerSchema{
			UniqueConstraints: []*UniqueConstraint{
				{Elements: []string{"ip", "port"}},
			},
		}
		got := cs.GetUniqueConstraints()
		if len(got) != 1 {
			t.Fatalf("GetUniqueConstraints() len = %d, want 1", len(got))
		}
		elems := got[0].GetElements()
		if len(elems) != 2 || elems[0] != "ip" || elems[1] != "port" {
			t.Errorf("GetElements() = %v, want [ip port]", elems)
		}
	})

	t.Run("multiple independent unique constraints", func(t *testing.T) {
		cs := &ContainerSchema{
			UniqueConstraints: []*UniqueConstraint{
				{Elements: []string{"ip", "port"}},
				{Elements: []string{"name"}},
			},
		}
		got := cs.GetUniqueConstraints()
		if len(got) != 2 {
			t.Fatalf("GetUniqueConstraints() len = %d, want 2", len(got))
		}
		if got[0].GetElements()[0] != "ip" {
			t.Errorf("constraint[0].elements[0] = %q, want %q", got[0].GetElements()[0], "ip")
		}
		if got[1].GetElements()[0] != "name" {
			t.Errorf("constraint[1].elements[0] = %q, want %q", got[1].GetElements()[0], "name")
		}
	})
}
