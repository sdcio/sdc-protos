package schema_server

func (c *ChoiceInfo) GetChoiceByName(choiceName string) *ChoiceInfoChoice {
	return c.Choice[choiceName]
}
