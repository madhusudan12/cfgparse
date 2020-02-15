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
	val := config.Get("newsection2", "ggveg")
	if val != "fjrhbfr" {
		t.Error("unable to get the value", val)
	}
}


func TestCfgParser_Get(t *testing.T) {
	config := New()
	err := config.ReadFile("config.ini")
	if err != nil {
		panic("Hey error while parsing file")
	}
	val := config.Get("newsection", "abc")
	if val != "newvalue" {
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
	config.Set("newsection", "newkey2", "newvalue2")
	val := config.Get("newsection2", "ggveg")
	if val != "fjrhbfr" {
		t.Error("unable to get the value", val)
	}
	config.Set("newsection2", "sea1", "animal")

	err = config.AddSection("newsection3")
	if err != nil {
		t.Error("Error adding section", err)
	}
}