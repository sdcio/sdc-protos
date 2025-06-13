package schema_server

import (
	"cmp"
	"math"
	"strconv"
	"strings"
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

// ToString converts the TypedValue to the real, non proto string
func (tv *TypedValue) ToString() string {
	switch tv.Value.(type) {
	case *TypedValue_AnyVal:
		return string(tv.GetAnyVal().GetValue()) // questionable...
	case *TypedValue_AsciiVal:
		return tv.GetAsciiVal()
	case *TypedValue_BoolVal:
		return strconv.FormatBool(tv.GetBoolVal())
	case *TypedValue_BytesVal:
		return string(tv.GetBytesVal()) // questionable...
	case *TypedValue_DecimalVal:
		d := tv.GetDecimalVal()
		digitsStr := strconv.FormatInt(d.Digits, 10)
		negative := false
		if d.Digits < 0 {
			negative = true
			digitsStr = digitsStr[1:] // Remove the "-" sign for processing
		}
		// Add leading zeros if necessary
		for uint32(len(digitsStr)) <= d.Precision {
			digitsStr = "0" + digitsStr
		}
		// Insert the decimal point
		if d.Precision > 0 {
			decimalPointIndex := len(digitsStr) - int(d.Precision)
			digitsStr = digitsStr[:decimalPointIndex] + "." + digitsStr[decimalPointIndex:]
		}
		// Add back the negative sign if necessary
		if negative {
			digitsStr = "-" + digitsStr
		}
		return digitsStr
	case *TypedValue_DoubleVal:
		return strconv.FormatFloat(tv.GetDoubleVal(), byte('e'), -1, 64)
	case *TypedValue_EmptyVal:
		return "{}"
	case *TypedValue_FloatVal:
		return strconv.FormatFloat(float64(tv.GetFloatVal()), byte('e'), -1, 64)
	case *TypedValue_IntVal:
		return strconv.Itoa(int(tv.GetIntVal()))
	case *TypedValue_JsonIetfVal:
		return string(tv.GetJsonIetfVal())
	case *TypedValue_JsonVal:
		return string(tv.GetJsonVal())
	case *TypedValue_LeaflistVal:
		rs := make([]string, 0, len(tv.GetLeaflistVal().GetElement()))
		for _, lfv := range tv.GetLeaflistVal().GetElement() {
			rs = append(rs, lfv.ToString())
		}
		return strings.Join(rs, ",")
	case *TypedValue_ProtoBytes:
		return string(tv.GetProtoBytes()) // questionable
	case *TypedValue_StringVal:
		return tv.GetStringVal()
	case *TypedValue_UintVal:
		return strconv.Itoa(int(tv.GetUintVal()))
	case *TypedValue_IdentityrefVal:
		return tv.GetIdentityrefVal().Value
	}
	return ""
}
