package sdcpb

func (t *TransactionSetResponseIntent) Failed() bool {
	return len(t.Errors) > 0
}
