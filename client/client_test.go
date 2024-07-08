package main

import (
	"rcunov/net-transfer/utils"
	"testing"
)

func TestLoadCert(t *testing.T) {
	_, err := utils.LoadCert(certFile, keyFile)
	if err != nil {
		t.Errorf("could not load cert/key at %v and %v. error: %v", certFile, keyFile, err)
	}
}