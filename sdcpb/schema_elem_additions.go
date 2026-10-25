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

// IsSensitive reports whether the SchemaElem marks the node as sensitive.
// Returns false for containers and nil receivers.
func (s *SchemaElem) IsSensitive() bool {
	if s == nil {
		return false
	}
	switch x := s.Schema.(type) {
	case *SchemaElem_Field:
		return x.Field.GetSensitive()
	case *SchemaElem_Leaflist:
		return x.Leaflist.GetSensitive()
	}
	return false
}
