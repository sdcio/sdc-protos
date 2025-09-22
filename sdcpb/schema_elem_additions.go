package schema_server

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
