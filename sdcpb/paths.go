package sdcpb

type Paths []*Path

func (p Paths) ToXPathSlice() []string {
	result := make([]string, 0, len(p))
	for _, x := range p {
		result = append(result, x.ToXPath(false))
	}
	return result
}
