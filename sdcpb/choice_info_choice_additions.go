package sdcpb

func (c *ChoiceInfoChoice) GetAllAttributes() []string {
	result := make([]string, 0, len(c.Case))

	for _, cas := range c.GetCase() {
		result = append(result, cas.GetElements()...)
	}
	return result
}
