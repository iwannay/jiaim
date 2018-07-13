package main

import (
	"testing"
)

func TestRing_Get(t *testing.T) {

	r := NewRing(100)

	v, err := r.Set()
	if err != nil {
		t.Fatal(err)
	}
	v.MsgId = 1
	t.Log("set:", v)
	r.SetAdv()

	v, err = r.Get()
	if err != nil {
		t.Fatal(err)
	}
	r.GetAdv()
	t.Log("get", v)

	v, err = r.Get()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(v)

}
