package sdcpb

import "testing"

func BenchmarkConvertString(b *testing.B) {
	lst := &SchemaLeafType{
		Type: "string",
		Patterns: []*SchemaPattern{
			{Pattern: `[a-zA-Z0-9\-_.]+`},
			{Pattern: `.*\.tmp`, Inverted: true},
		},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ConvertString("interface-eth0", lst); err != nil {
			b.Fatal(err)
		}
	}
}
