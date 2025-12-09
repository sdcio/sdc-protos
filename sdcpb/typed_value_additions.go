package sdcpb

import (
	"bytes"
	"cmp"
	"fmt"
	"math"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (tv *TypedValue) typeOrder() int {
	msg := tv.ProtoReflect()
	fd := msg.Descriptor().Oneofs().Get(0) // the oneof
	which := msg.WhichOneof(fd)            // active field
	return int(which.Number())             // field number = stable ordering
}

// Equal provide equal via Cmp
func (tv *TypedValue) Equal(other *TypedValue) bool {
	return tv.Cmp(other) == 0
}

// Cmp implements the slices.SortFunc for TypedValues
// if the two TypedValues are of different TypedValues types, 0 is returned (indicating equal), since no proper comparison is possible
func (tv *TypedValue) Cmp(other *TypedValue) int {
	if tv == nil && other == nil {
		return 0
	}
	if tv == nil {
		return -1
	}
	if other == nil {
		return 1
	}
	// if types are different, we use the rank from the proto file to determine order.
	if reflect.TypeOf(tv.GetValue()) != reflect.TypeOf(other.GetValue()) {
		return cmp.Compare(tv.typeOrder(), other.typeOrder())
	}

	// otherwise compare the exact types values.
	switch tv.Value.(type) {
	case *TypedValue_AnyVal:
		return bytes.Compare(tv.GetAnyVal().Value, other.GetAnyVal().GetValue())
	case *TypedValue_AsciiVal:
		return cmp.Compare(tv.GetAsciiVal(), other.GetAsciiVal())
	case *TypedValue_BoolVal:
		return cmp.Compare(boolToInt(tv.GetBoolVal()), boolToInt(other.GetBoolVal()))
	case *TypedValue_BytesVal:
		return bytes.Compare(tv.GetBytesVal(), other.GetBytesVal())
	case *TypedValue_DecimalVal:
		dtv := tv.GetDecimalVal()
		dother := other.GetDecimalVal()

		// pick the higher precision
		maxPrec := int(dtv.Precision)
		if int(dother.Precision) > maxPrec {
			maxPrec = int(dother.Precision)
		}

		// rescale digits to the same base
		scaletv := int(math.Pow10(maxPrec - int(dtv.Precision)))
		scaleother := int(math.Pow10(maxPrec - int(dother.Precision)))

		ntv := dtv.Digits * int64(scaletv)
		nother := dother.Digits * int64(scaleother)

		return cmp.Compare(ntv, nother)
	case *TypedValue_DoubleVal:
		return cmp.Compare(tv.GetDoubleVal(), other.GetDoubleVal())
	case *TypedValue_EmptyVal:
		return 0
	case *TypedValue_FloatVal:
		return cmp.Compare(tv.GetFloatVal(), other.GetFloatVal())
	case *TypedValue_IntVal:
		return cmp.Compare(tv.GetIntVal(), other.GetIntVal())
	case *TypedValue_JsonIetfVal:
		return bytes.Compare(tv.GetJsonIetfVal(), other.GetJsonIetfVal())
	case *TypedValue_JsonVal:
		return bytes.Compare(tv.GetJsonVal(), other.GetJsonVal())
	case *TypedValue_LeaflistVal:
		lltv := toStringSorted(tv.GetLeaflistVal().GetElement())
		llother := toStringSorted(other.GetLeaflistVal().GetElement())
		return slices.Compare(lltv, llother)
	case *TypedValue_ProtoBytes:
		return bytes.Compare(tv.GetProtoBytes(), other.GetProtoBytes())
	case *TypedValue_StringVal:
		return cmp.Compare(tv.GetStringVal(), other.GetStringVal())
	case *TypedValue_UintVal:
		return cmp.Compare(tv.GetUintVal(), other.GetUintVal())
	case *TypedValue_IdentityrefVal:
		tvVal := fmt.Sprintf("%s%s%s", tv.GetIdentityrefVal().GetValue(), tv.GetIdentityrefVal().GetModule(), tv.GetIdentityrefVal().GetPrefix())
		otherVal := fmt.Sprintf("%s%s%s", other.GetIdentityrefVal().GetValue(), other.GetIdentityrefVal().GetModule(), other.GetIdentityrefVal().GetPrefix())
		return cmp.Compare(tvVal, otherVal)
	}
	return 0
}

// toStringSorted takes a slice of TVs converts the elements to strings and returns a the sorted string slice.
func toStringSorted(tvs []*TypedValue) []string {
	result := make([]string, 0, len(tvs))
	for _, tv := range tvs {
		result = append(result, tv.ToString())
	}
	slices.Sort(result)
	return result
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
		return strconv.FormatInt(tv.GetIntVal(), 10)
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
		return strconv.FormatUint(tv.GetUintVal(), 10)
	case *TypedValue_IdentityrefVal:
		return tv.GetIdentityrefVal().Value
	}
	return ""
}

//TODO: Still needed?
func ConvertToTypedValue(schemaObject *SchemaElem, v string, ts uint64) (*TypedValue, error) {
	var schemaType *SchemaLeafType
	switch {
	case schemaObject.GetField() != nil:
		schemaType = schemaObject.GetField().GetType()
	case schemaObject.GetLeaflist() != nil:
		schemaType = schemaObject.GetLeaflist().GetType()
	case schemaObject.GetContainer() != nil:
		if !schemaObject.GetContainer().IsPresence {
			return nil, fmt.Errorf("non presence container update")
		}
		return nil, nil
	}
	return TVFromString(schemaType, v, ts)
}

func (tv *TypedValue) ToYANGType(schemaObject *SchemaElem) (*TypedValue, error) {
	switch tv.Value.(type) {
	case *TypedValue_AsciiVal:
		return ConvertToTypedValue(schemaObject, tv.GetAsciiVal(), tv.GetTimestamp())
	case *TypedValue_BoolVal:
		return tv, nil
	case *TypedValue_BytesVal:
		return tv, nil
	case *TypedValue_DecimalVal:
		return tv, nil
	case *TypedValue_FloatVal:
		return tv, nil
	case *TypedValue_DoubleVal:
		return tv, nil
	case *TypedValue_IntVal:
		return tv, nil
	case *TypedValue_StringVal:
		return ConvertToTypedValue(schemaObject, tv.GetStringVal(), tv.GetTimestamp())
	case *TypedValue_UintVal:
		return tv, nil
	case *TypedValue_JsonIetfVal: // TODO:
	case *TypedValue_JsonVal: // TODO:
	case *TypedValue_LeaflistVal:
		return tv, nil
	case *TypedValue_ProtoBytes:
		return tv, nil
	case *TypedValue_AnyVal:
		return tv, nil
	case *TypedValue_IdentityrefVal:
		return ConvertToTypedValue(schemaObject, tv.GetStringVal(), tv.GetTimestamp())
	}
	return tv, nil
}
