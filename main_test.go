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

func TestCalculateDynamicPath(t *testing.T) {

	g := NewGraph()
	g.AddRate("USD", "BTC", 1.0)
	g.AddRate("BTC", "ETH", 1.0)
	g.AddRate("ETH", "USD", 1.1)

	res := CalculateDynamicPath(g, 100, "USD", 100, []string{"USD"}, 4)
	expected := []string{"USD", "BTC", "ETH", "USD"}

	if !reflect.DeepEqual(res[0].Path, expected) {
		t.Errorf("Wrong path returned. Expected %v, got %v", expected, res[0].Path)
	}

	g = NewGraph()
	g.AddRate("USD", "BTC", 1.0)
	g.AddRate("BTC", "ETH", 1.0)
	g.AddRate("ETH", "USD", 0.9)

	res = CalculateDynamicPath(g, 100, "USD", 100, []string{"USD"}, 4)
	expected = []string{}

	if len(res) != 0 {
		t.Errorf("Wrong path returned. Expected %v, got %v", expected, res[0].Path)
	}

}
