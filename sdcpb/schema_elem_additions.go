package sdcpb

func (s *SchemaElem) IsState() bool {
	switch x := s.Schema.(type) {
	case *SchemaElem_Container:
		return x.Container.IsState
	case *SchemaElem_Field:
		return x.Field.IsState
	case *SchemaElem_Leaflist:
		return x.Leaflist.IsState
	}
	return false
}

// IsMutateOnWrite reports whether s is a leaf or leaf-list whose non-empty
// readback value on the same path after a write may differ from the value
// written. Containers and nil or unset elements return false.
func (s *SchemaElem) IsMutateOnWrite() bool {
	if s == nil {
		return false
	}
	switch x := s.Schema.(type) {
	case *SchemaElem_Field:
		return x.Field.GetMutateOnWrite()
	case *SchemaElem_Leaflist:
		return x.Leaflist.GetMutateOnWrite()
	}
	return false
}
