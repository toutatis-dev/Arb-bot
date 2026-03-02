package main

import (
	"reflect"
	"testing"
)

func TestNewGraph(t *testing.T) {

	new := NewGraph()
	expectedGraph := &Graph{
		Rates: make(map[string]map[string]float64),
	}

	if !reflect.DeepEqual(new, expectedGraph) {
		t.Errorf("NewGraph did not return an empy Graph struct. Expected %v, got %v", expectedGraph, new)
	}

}

func TestAddRate(t *testing.T) {
	g := NewGraph()
	g.AddRate("USD", "BTC", 65000)
	expected := map[string]map[string]float64{"USD": {"BTC": 65000}}

	if !reflect.DeepEqual(g.Rates, expected) {
		t.Errorf("AddRate failed. Expected %v, got %v", expected, g.Rates)
	}

}
