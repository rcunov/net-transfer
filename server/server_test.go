package main

import (
	"rcunov/net-transfer/utils"
	"testing"
)

func TestLoadCert(t *testing.T) {
	_, err := utils.LoadCert(certFile, keyFile)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestStartServer(t *testing.T) {
	_, err := StartServer(port, certFile, keyFile)
	if err != nil {
		t.Errorf(err.Error())
	}
}