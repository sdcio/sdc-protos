package sdcpb

func (x *ContainerSchema) GetMandatoryChildrenConfig() []*MandatoryChild {
	var result []*MandatoryChild
	if x != nil {
		for _, c := range x.MandatoryChildren {
			if !c.GetIsState() {
				result = append(result, c)
			}
		}
	}
	return result
}

func (x *ContainerSchema) GetMandatoryChildrenState() []*MandatoryChild {
	var result []*MandatoryChild
	if x != nil {
		for _, c := range x.MandatoryChildren {
			if c.GetIsState() {
				result = append(result, c)
			}
		}
	}
	return result
}
