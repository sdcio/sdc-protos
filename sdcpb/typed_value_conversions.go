package sdcpb

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	logf "github.com/sdcio/logger"
	"github.com/sdcio/sdc-protos/utils"
)

func TVFromString(schemaType *SchemaLeafType, v string, ts uint64) (*TypedValue, error) {
	if schemaType == nil {
		return nil, fmt.Errorf("schemaType cannot be nil")
	}

	var tv *TypedValue
	var err error
	switch schemaType.Type {
	case "string":
		tv, err = ConvertString(v, schemaType)
	case "union":
		tv, err = ConvertUnion(v, schemaType.UnionTypes)
	case "boolean":
		tv, err = ConvertBoolean(v, schemaType)
	case "int8":
		// TODO: HEX and OCTAL pre-processing for all INT types
		// https://www.rfc-editor.org/rfc/rfc6020.html#page-112
		tv, err = ConvertInt8(v, schemaType)
	case "int16":
		tv, err = ConvertInt16(v, schemaType)
	case "int32":
		tv, err = ConvertInt32(v, schemaType)
	case "int64":
		tv, err = ConvertInt64(v, schemaType)
	case "uint8":
		tv, err = ConvertUint8(v, schemaType)
	case "uint16":
		tv, err = ConvertUint16(v, schemaType)
	case "uint32":
		tv, err = ConvertUint32(v, schemaType)
	case "uint64":
		tv, err = ConvertUint64(v, schemaType)
	case "enumeration":
		tv, err = ConvertEnumeration(v, schemaType)
	case "empty":
		tv, err = &TypedValue{Value: &TypedValue_EmptyVal{}}, nil
	case "bits":
		tv, err = ConvertBits(v, schemaType)
	case "binary": // https://www.rfc-editor.org/rfc/rfc6020.html#section-9.8
		tv, err = ConvertBinary(v, schemaType)
	case "leafref": // https://www.rfc-editor.org/rfc/rfc6020.html#section-9.9
		tv, err = ConvertLeafRef(v, schemaType)
	case "identityref": //TODO: https://www.rfc-editor.org/rfc/rfc6020.html#section-9.10
		tv, err = ConvertIdentityRef(v, schemaType)
	case "instance-identifier": //TODO: https://www.rfc-editor.org/rfc/rfc6020.html#section-9.13
		tv, err = ConvertInstanceIdentifier(v, schemaType)
	case "decimal64":
		// TODO: is the following TODO still valid? I think no
		// TODO: fraction-digits (https://www.rfc-editor.org/rfc/rfc6020.html#section-9.3.4)
		tv, err = ConvertDecimal64(v, schemaType)
	default:
		tv, err = nil, fmt.Errorf("FromString conversion not implemented for type '%s'", schemaType.Type)
	}

	if err != nil {
		return nil, err
	}
	// Set timestamp
	tv.Timestamp = ts
	return tv, nil
}

func ConvertInstanceIdentifier(value string, slt *SchemaLeafType) (*TypedValue, error) {
	// delegate to string, validation is left for a different party at a later stage in processing
	return ConvertString(value, slt)
}

func ConvertIdentityRef(value string, schemaType *SchemaLeafType) (*TypedValue, error) {
	before, name, found := strings.Cut(value, ":")
	if !found {
		name = before
	}
	prefix, ok := schemaType.IdentityPrefixesMap[name]
	if !ok {
		identities := make([]string, 0, len(schemaType.IdentityPrefixesMap))
		for k := range schemaType.IdentityPrefixesMap {
			identities = append(identities, k)
		}
		return nil, fmt.Errorf("identity %s not found, possible values are %s", value, strings.Join(identities, ", "))
	}
	module, ok := schemaType.ModulePrefixMap[name]
	if !ok {
		identities := make([]string, 0, len(schemaType.IdentityPrefixesMap))
		for k := range schemaType.IdentityPrefixesMap {
			identities = append(identities, k)
		}
		return nil, fmt.Errorf("identity %s not found, possible values are %s", value, strings.Join(identities, ", "))
	}
	return &TypedValue{
		Value: &TypedValue_IdentityrefVal{IdentityrefVal: &IdentityRef{Value: name, Prefix: prefix, Module: module}},
	}, nil
}

func ConvertBinary(value string, slt *SchemaLeafType) (*TypedValue, error) {
	// Binary is basically a base64 encoded string that might carry a length restriction
	// so we should be fine with delegating to string
	return ConvertString(value, slt)
}

func ConvertLeafRef(value string, slt *SchemaLeafType) (*TypedValue, error) {
	// Try to convert based on the target type info
	return TVFromString(slt.LeafrefTargetType, value, 0)
}

func ConvertEnumeration(value string, slt *SchemaLeafType) (*TypedValue, error) {
	// iterate the valid values as per schema
	for _, item := range slt.EnumNames {
		// if value is found, return a StringVal
		if value == item {
			return &TypedValue{
				Value: &TypedValue_StringVal{
					StringVal: value,
				},
			}, nil
		}
	}
	// If value is not found return an error
	return nil, fmt.Errorf("value %q does not match any valid enum values [%s]", value, strings.Join(slt.EnumNames, ", "))
}

func ConvertBoolean(value string, _ *SchemaLeafType) (*TypedValue, error) {
	bval, err := strconv.ParseBool(value)
	if err != nil {
		// if it is any other value, return error
		return nil, err
	}
	// otherwise return the BoolVal TypedValue
	return &TypedValue{
		Value: &TypedValue_BoolVal{
			BoolVal: bval,
		},
	}, nil
}

func ConvertSdcpbNumberToUint64(mm *Number) (uint64, error) {
	if mm.Negative {
		return 0, fmt.Errorf("negative number to uint conversion")
	}
	return mm.Value, nil
}

func intAbs(x int64) uint64 {
	ui := uint64(x)
	if x < 0 {
		return ^(ui) + 1
	}
	return ui
}

func ConvertSdcpbNumberToInt64(mm *Number) (int64, error) {
	if mm.Negative {
		if mm.Value > intAbs(math.MinInt64) {
			return 0, fmt.Errorf("error converting -%d to int64: overflow", mm.Value)
		}
		return -int64(mm.Value), nil
	}

	if mm.Value > math.MaxInt64 {
		return 0, fmt.Errorf("error converting %d to int64 overflow", mm.Value)
	}
	return int64(mm.Value), nil
}

func convertUint(value string, minMaxs []*SchemaMinMaxType, ranges *utils.Rnges[uint64]) (*TypedValue, error) {
	if ranges == nil {
		ranges = utils.NewRnges[uint64]()
	}
	for _, x := range minMaxs {
		min, err := ConvertSdcpbNumberToUint64(x.Min)
		if err != nil {
			return nil, err
		}
		max, err := ConvertSdcpbNumberToUint64(x.Max)
		if err != nil {
			return nil, err
		}
		ranges.AddRange(min, max)
	}

	uValue, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return nil, err
	}
	// validate the value against the ranges
	valid := ranges.IsWithinAnyRange(uValue)
	if !valid {
		return nil, fmt.Errorf("%q not within ranges: %s", value, ranges.String())
	}
	// return the TypedValue
	return &TypedValue{Value: &TypedValue_UintVal{UintVal: uValue}}, nil
}

func ConvertUint8(value string, lst *SchemaLeafType) (*TypedValue, error) {
	// create the ranges
	ranges := utils.NewRnges[uint64]()
	ranges.AddRange(0, math.MaxUint8)

	return convertUint(value, lst.Range, ranges)
}

func ConvertUint16(value string, lst *SchemaLeafType) (*TypedValue, error) {
	// create the ranges
	ranges := utils.NewRnges[uint64]()
	ranges.AddRange(0, math.MaxUint16)

	return convertUint(value, lst.Range, ranges)
}

func ConvertUint32(value string, lst *SchemaLeafType) (*TypedValue, error) {
	// create the ranges
	ranges := utils.NewRnges[uint64]()
	ranges.AddRange(0, math.MaxUint32)

	return convertUint(value, lst.Range, ranges)
}

func ConvertUint64(value string, lst *SchemaLeafType) (*TypedValue, error) {
	// create the ranges
	ranges := utils.NewRnges[uint64]()

	return convertUint(value, lst.Range, ranges)
}

func convertInt(value string, minMaxs []*SchemaMinMaxType, ranges *utils.Rnges[int64]) (*TypedValue, error) {
	for _, x := range minMaxs {
		min, err := ConvertSdcpbNumberToInt64(x.Min)
		if err != nil {
			return nil, err
		}
		max, err := ConvertSdcpbNumberToInt64(x.Max)
		if err != nil {
			return nil, err
		}
		ranges.AddRange(min, max)
	}

	// validate the value against the ranges
	iValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, err
	}
	// validate the value against the ranges
	valid := ranges.IsWithinAnyRange(iValue)
	if !valid {
		return nil, fmt.Errorf("%q not within ranges: %s", value, ranges.String())
	}
	// return the TypedValue
	return &TypedValue{Value: &TypedValue_IntVal{IntVal: iValue}}, nil
}

func ConvertInt8(value string, lst *SchemaLeafType) (*TypedValue, error) {
	// create the ranges
	ranges := utils.NewRnges[int64]()
	ranges.AddRange(math.MinInt8, math.MaxInt8)

	return convertInt(value, lst.Range, ranges)
}

func ConvertInt16(value string, lst *SchemaLeafType) (*TypedValue, error) {
	// create the ranges
	ranges := utils.NewRnges[int64]()
	ranges.AddRange(math.MinInt16, math.MaxInt16)

	return convertInt(value, lst.Range, ranges)
}

func ConvertInt32(value string, lst *SchemaLeafType) (*TypedValue, error) {
	// create the ranges
	ranges := utils.NewRnges[int64]()
	ranges.AddRange(math.MinInt32, math.MaxInt32)

	return convertInt(value, lst.Range, ranges)
}
func ConvertInt64(value string, lst *SchemaLeafType) (*TypedValue, error) {
	// create the ranges
	ranges := utils.NewRnges[int64]()

	return convertInt(value, lst.Range, ranges)
}

func xmlRegexConvert(s string) string {

	cTest := func(r rune, prev rune) bool {
		// if ^ is not following a [ or if $ we want to return true
		return (r == '^' && prev != '[') || r == '$'
	}

	b := strings.Builder{}
	b.Grow(len(s) + len(s)/4)
	slashes := 0
	prevR := rune(0)

	for _, r := range s {
		if r == '\\' {
			slashes++
			prevR = r
			b.WriteRune(r)
			continue
		}

		if cTest(r, prevR) && slashes%2 == 0 {
			b.WriteRune('\\')
		}

		slashes = 0
		prevR = r
		b.WriteRune(r)
	}
	return b.String()
}

func ConvertString(value string, lst *SchemaLeafType) (*TypedValue, error) {
	// check length of the string if the length property is set
	// length will contain a range like string definition "5..60" or "7..10|40..45"
	if len(lst.Length) != 0 {
		_, err := convertUint(strconv.Itoa(len(value)), lst.Length, nil)

		if err != nil {
			return nil, err
		}

	}

	overallMatch := true
	// If the type has multiple "pattern" statements, the expressions are
	// ANDed together, i.e., all such expressions have to match.
	for _, sp := range lst.Patterns {
		// The set of metacharacters is not the same between XML schema and perl/python/go REs
		// the set of metacharacters for XML is: .\?*+{}()[] (https://www.w3.org/TR/xmlschema-2/#dt-metac)
		// the set of metacharacters defined in go is: \.+*?()|[]{}^$ (go/libexec/src/regexp/regexp.go:714)
		// we need therefore to escape some values
		// TODO check about '^'

		escaped := xmlRegexConvert(sp.Pattern)
		re, err := regexp.Compile(escaped)
		if err != nil {
			//TODO: Do we want to stop here?
			logf.DefaultLogger.Error(err, "unable to compile regex", "pattern", sp.Pattern)
			return nil, fmt.Errorf("unable to compile regex: %w", err)
		}
		match := re.MatchString(value)
		// if it is a match and not inverted
		// or it is not a match but inverted
		// then this is valid
		if (match && !sp.Inverted) || (!match && sp.Inverted) {
			continue
		} else {
			overallMatch = false
			break
		}
	}
	if overallMatch {
		return &TypedValue{
			Value: &TypedValue_StringVal{
				StringVal: value,
			},
		}, nil
	}
	return nil, fmt.Errorf("%q does not match patterns", value)

}

func ConvertDecimal64(value string, lst *SchemaLeafType) (*TypedValue, error) {
	v := strings.TrimSpace(value)
	if v == "" {
		return nil, fmt.Errorf("empty decimal64 string")
	}

	neg := false
	if strings.HasPrefix(v, "-") {
		neg = true
		v = v[1:]
	} else if strings.HasPrefix(v, "+") {
		v = v[1:]
	}

	if v == "" {
		return nil, fmt.Errorf("no digits after sign")
	}

	parts := strings.SplitN(v, ".", 2)
	intPart := parts[0]
	fracPart := ""
	if len(parts) == 2 {
		fracPart = parts[1]
	}

	// Require at least one digit total (either int or frac)
	if intPart == "" && fracPart == "" {
		return nil, fmt.Errorf("no digits in decimal64 value")
	}

	if intPart == "" {
		intPart = "0"
	}

	combined := intPart + fracPart
	if combined == "" {
		return nil, fmt.Errorf("no digits to parse")
	}

	digits, err := strconv.ParseInt(combined, 10, 64)
	if err != nil {
		return nil, err
	}
	if neg {
		digits = -digits
	}

	precision := uint32(len(fracPart))

	d64 := &Decimal64{
		Digits:    digits,
		Precision: precision,
	}

	return &TypedValue{
		Value: &TypedValue_DecimalVal{
			DecimalVal: d64,
		},
	}, nil
}

func ConvertUnion(value string, slts []*SchemaLeafType) (*TypedValue, error) {
	// iterate over the union types try to convert without error
	for _, slt := range slts {
		tv, err := TVFromString(slt, value, 0)
		// if no error type conversion was fine
		if err != nil {
			continue
		}
		// return the TypedValue
		return tv, nil
	}
	return nil, fmt.Errorf("no union type fit the provided value %q", value)
}

func validateBitString(value string, allowed []*Bit) bool {
	//split string to individual bits
	bits := strings.Fields(value)
	// empty string is fine
	if len(bits) == 0 {
		return true
	}
	// track pos inside allowed slice
	pos := 0
	for _, b := range bits {
		// increase pos until we get to an allowed bit or we reach the end of the slice
		for pos < len(allowed) && allowed[pos].GetName() != b {
			pos++
		}
		// if we are at the end of the array, we did not validate
		if pos == len(allowed) {
			return false
		}
		//move past found element
		pos++
	}
	return true
}

func ConvertBits(value string, slt *SchemaLeafType) (*TypedValue, error) {
	if slt == nil {
		return nil, fmt.Errorf("type information is nil")
	}
	if len(slt.Bits) == 0 {
		return nil, fmt.Errorf("type information is missing bits information")
	}
	if validateBitString(value, slt.Bits) {
		return &TypedValue{
			Value: &TypedValue_StringVal{
				StringVal: value,
			},
		}, nil
	}
	// If value is not valid return an error
	validBits := make([]string, 0, len(slt.Bits))
	for _, b := range slt.Bits {
		validBits = append(validBits, b.GetName())
	}
	return nil, fmt.Errorf("value %q does not follow required bit ordering [%s]", value, strings.Join(validBits, " "))
}
