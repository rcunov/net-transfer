package main

import "testing"

func TestExampleFunc(t *testing.T) {
	result := ExampleFunc()
	
	if result != "one" {
		t.Errorf("test function failed, got: %s, want %s", result, "one")
	}
}