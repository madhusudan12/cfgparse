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
	val := config.Get("base", "username")
	if val != "madhusudan" {
		t.Error("unable to get the value", val)
	}
}


func TestCfgParser_AddSection(t *testing.T) {
	config := New()
	err := config.ReadFile("config.ini")
	if err != nil {
		panic("Hey error while parsing file")
	}
	err = config.AddSection("newsection2")
	if err != nil {
		t.Error("Error adding section", err)
	}

}


func TestCfgParser_Set(t *testing.T) {
	config := New()
	err := config.ReadFile("config.ini")
	if err != nil {
		panic("Hey error while parsing file")
	}
	config.Set("newsection", "newkey", "newvalue")
}