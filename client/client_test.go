package main

import (
	"testing"
)

func TestGenerateCert(t *testing.T) {
	_, err := GenerateCert()
	if err != nil {
		t.Errorf("could not generate the certificate. error: %v", err)
	}
}