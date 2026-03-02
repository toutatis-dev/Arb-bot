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
