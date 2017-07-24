package id

import (
	"testing"
)

func Test_getFromString(t *testing.T) {

	idGenerate := newIdGenerate()
	for i := 0; i < 100; i++ {
		t.Run(string(i), func(t *testing.T) {
			got := idGenerate.string()
			t.Logf("self.string() = %s", got)
		})
	}
}

func Test_counter(t *testing.T) {
	tests := []struct {
		name string
		want int32
	}{
		struct {
			name string
			want int32
		}{name: "1"},
		struct {
			name string
			want int32
		}{name: "2"},
	}
	idGenerate := newIdGenerate()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := idGenerate.counter()
			t.Logf("self.counter() = %v", got)

		})
	}
}
