package schema_server

import (
	"cmp"
	"math"
)

// Cmp implements the slices.SortFunc for TypedValues
// if the two TypedValues are of different TypedValues types, 0 is returned (indicating equal), since no proper comparison is possible
func (tv *TypedValue) Cmp(other *TypedValue) int {
	switch tv.Value.(type) {
	case *TypedValue_BoolVal:
		_, ok := other.Value.(*TypedValue_BoolVal)
		if !ok {
			return 0
		}
		if tv.GetBoolVal() && !other.GetBoolVal() {
			return 1
		}
		if !tv.GetBoolVal() && other.GetBoolVal() {
			return -1
		}
	case *TypedValue_DecimalVal:
		_, ok := other.Value.(*TypedValue_DecimalVal)
		if !ok {
			return 0
		}
		tvScale := math.Pow(10, float64(tv.GetDecimalVal().Precision))
		otherScale := math.Pow(10, float64(other.GetDecimalVal().Precision))

		tvVal := tv.GetDecimalVal().Digits * int64(tvScale)
		otherVal := other.GetDecimalVal().Digits * int64(otherScale)
		// check digits are larger or smaller
		if tvVal > otherVal {
			return 1
		} else if tvVal < otherVal {
			return -1
		}

	case *TypedValue_DoubleVal:
		_, ok := other.Value.(*TypedValue_DoubleVal)
		if !ok {
			return 0
		}
		if tv.GetDoubleVal() > other.GetDoubleVal() {
			return 1
		} else if tv.GetDoubleVal() < other.GetDoubleVal() {
			return -1
		}
	case *TypedValue_FloatVal:
		_, ok := other.Value.(*TypedValue_FloatVal)
		if !ok {
			return 0
		}
		if tv.GetFloatVal() > other.GetFloatVal() {
			return 1
		} else if tv.GetFloatVal() < other.GetFloatVal() {
			return -1
		}
	case *TypedValue_IntVal:
		_, ok := other.Value.(*TypedValue_IntVal)
		if !ok {
			return 0
		}
		if tv.GetIntVal() > other.GetIntVal() {
			return 1
		} else if tv.GetIntVal() < other.GetIntVal() {
			return -1
		}
	case *TypedValue_UintVal:
		_, ok := other.Value.(*TypedValue_UintVal)
		if !ok {
			return 0
		}
		if tv.GetUintVal() > other.GetUintVal() {
			return 1
		} else if tv.GetUintVal() < other.GetUintVal() {
			return -1
		}
	case *TypedValue_StringVal:
		if other.GetStringVal() == "" {
			return 0
		}
		return cmp.Compare(tv.GetStringVal(), other.GetStringVal())
	case *TypedValue_IdentityrefVal:
		if other.GetIdentityrefVal() == nil {
			return 0
		}
		return cmp.Compare(tv.GetIdentityrefVal().Value, other.GetIdentityrefVal().Value)
	}
	return 0
}
