package specs

import "testing"

func BenchmarkValidateVersion(b *testing.B) {
	tests := []struct {
		doc     string
		version string
	}{
		{"first", "1.1.0"},
		{"middle", "0.6.0"},
		{"last", "0.1.0"},
		{"invalid", "2.0.0"},
	}

	for _, tc := range tests {
		b.Run(tc.doc, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = ValidateVersion(&Spec{Version: tc.version})
			}
		})
	}
}
