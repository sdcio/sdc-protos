package sdcpb

import (
	"fmt"
	"strings"
)

func ParseEncoding(enc string) (Encoding, error) {
	switch strings.ToLower(enc) {
	case "json":
		return Encoding_JSON, nil
	case "json_ietf":
		return Encoding_JSON_IETF, nil
	case "proto":
		return Encoding_PROTO, nil
	case "string":
		return Encoding_STRING, nil
	default:
		return 0, fmt.Errorf("unknown encoding: %s", enc)
	}
}
