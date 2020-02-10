package cfgparse

import (
	"testing"
)

func TestCfgParser_ReadFile(t *testing.T) {
	config := New()
	err := config.ReadFile("config.ini")
	if err != nil {
		panic("Hey error while parsing file")
	}
	val := config.Get("default", "usernme")
	if val != "madhusudan" {
		t.Error("unable to get the value", val)
	}
}
