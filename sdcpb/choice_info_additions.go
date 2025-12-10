package sdcpb

func (c *ChoiceInfo) GetChoiceByName(choiceName string) *ChoiceInfoChoice {
	return c.Choice[choiceName]
}
