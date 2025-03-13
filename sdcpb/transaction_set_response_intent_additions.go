package schema_server

func (t *TransactionSetResponseIntent) Failed() bool {
	return len(t.Errors) > 0
}
