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
	var val string
	err := config.ReadFile("config.ini")
	if err != nil {
		panic("Hey error while parsing file")
	}
	config.Set("newsection", "newkey", "newvalue")
	val = config.Get("newsection", "newkey")
	if val != "newvalue" {
		t.Error("unable to get the value", val)
	}
	config.Set("addsection", "sea1", "animal")
	val = config.Get("addsection", "sea1")
	if val != "animal" {
		t.Error("unable to get the value", val)
	}
}