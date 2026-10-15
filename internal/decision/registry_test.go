package decision

import (
	"encoding/json"
	"testing"
)

type testFactory struct{ id string }

func (f testFactory) ID() string { return f.id }
func (testFactory) New(json.RawMessage) (Evaluator, error) {
	return nil, nil
}

func TestLoadRegistry(t *testing.T) {
	registry, err := Load(testFactory{id: "typesafe"})
	if err != nil {
		t.Fatal(err)
	}
	provider, ok := registry.GetProvider("typesafe")
	if !ok || provider.Name != "TypeSafe" {
		t.Fatalf("provider = %#v, found = %t", provider, ok)
	}
	model, ok := registry.GetModel("typesafe", "jev-1.14.0")
	if !ok || model.Name != "Jev 1.14" {
		t.Fatalf("model = %#v, found = %t", model, ok)
	}
	if _, ok := registry.GetModel("typesafe", "jev-1.13.0"); !ok {
		t.Fatal("previous TypeSafe model should remain supported")
	}
	provider, ok = registry.GetProvider("liquid")
	if !ok || provider.Name != "Liquid AI" {
		t.Fatalf("provider = %#v, found = %t", provider, ok)
	}
	model, ok = registry.GetModel("liquid", "d1-ultra")
	if !ok || model.Name != "D1 Ultra" {
		t.Fatalf("model = %#v, found = %t", model, ok)
	}
	if _, ok := registry.GetModel("liquid", "d1"); !ok {
		t.Fatal("existing Liquid model should remain supported")
	}
}

func TestLoadRegistryRejectsUnknownFactory(t *testing.T) {
	if _, err := Load(testFactory{id: "unknown"}); err == nil {
		t.Fatal("expected unknown factory error")
	}
}
