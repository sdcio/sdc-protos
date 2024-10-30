package schema_server

import "fmt"

// YangString resturns the Value as "<prefix>:<value>" as represented in the schema
func (x *IdentityRef) YangString() string {
	if x == nil {
		return ""
	}
	return fmt.Sprintf("%s:%s", x.Prefix, x.Value)
}

// JsonString resturns the Value as "<module>:<value>" as represented in json_ietf
func (x *IdentityRef) JsonIetfString() string {
	if x == nil {
		return ""
	}
	return fmt.Sprintf("%s:%s", x.Module, x.Value)
}
