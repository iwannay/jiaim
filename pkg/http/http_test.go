package http

import (
	"testing"
)

func TestHttpClient_Post(t *testing.T) {

	h := NewHttpClient()
	got, err := h.Post("http://dev.api.pay.verystar.cn", ContentTypeJSON, false, map[string]interface{}{
		"hello": "boy",
	})

	if err != nil {
		t.Error(err)
	}

	t.Log(string(got))
}

func BenchmarkPost(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		h := NewHttpClient()
		for pb.Next() {
			_, err := h.Post("http://dev.api.pay.verystar.cn", ContentTypeJSON, false, map[string]interface{}{
				"hello": "boy",
			})
			if err != nil {
				b.Error(err)
			}
		}

	})
	b.ReportAllocs()
}
