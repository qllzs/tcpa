package main

import "testing"

func TestDb(t *testing.T) {

	ret := QueryTcparByUeIP("30.0.11.11")
	if ret {
		t.Log("tcpar on")
	} else {
		t.Log("tcpar off")
	}

}
