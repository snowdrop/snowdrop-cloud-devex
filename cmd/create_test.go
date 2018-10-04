package cmd

import (
	"github.com/posener/complete"
	"reflect"
	"testing"
)

var moduleNames = []string{"cloud-kubernetes", "core", "integration", "jaxrs", "jpa", "keycloak", "monitoring", "web", "websocket"}

func TestModuleSuggester_Predict(t *testing.T) {
	suggester := moduleSuggester{}

	predict := suggester.Predict(complete.Args{Last: ""})
	check(predict, moduleNames, t)

	predict = suggester.Predict(complete.Args{Last: "w"})
	check(predict, moduleNames, t)

	predict = suggester.Predict(complete.Args{Last: "web,"})
	check(predict, []string{"web,cloud-kubernetes", "web,core", "web,integration", "web,jaxrs", "web,jpa", "web,keycloak", "web,monitoring", "web,websocket"}, t)

	predict = suggester.Predict(complete.Args{Last: "web,core,"})
	check(predict, []string{"web,core,cloud-kubernetes", "web,core,integration", "web,core,jaxrs", "web,core,jpa", "web,core,keycloak", "web,core,monitoring", "web,core,websocket"}, t)
}

func check(predict []string, expected []string, t *testing.T) {
	if !reflect.DeepEqual(predict, expected) {
		t.Errorf("Expected %s, got %s", expected, predict)
	}
}
