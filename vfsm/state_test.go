package vfsm

import "testing"

type Sample struct {
}

func (s Sample) IsStopped() bool {
	return false
}
func (s Sample) Error() error {
	return nil
}
func (s Sample) Descriptor() StateObjectDescriptor {
	return StateObjectDescriptor{
		"Sample", "sample object", []StateDescriptor{{"a", "a", []string{"b"}}, {"b", "b", nil}},
	}
}
func (s Sample) Current() string {
	return "a"
}

func (s Sample) Shutdown() error {
	return nil
}

func (s Sample) Start() error {
	return nil
}
func TestState(t *testing.T) {
	s1 := State[Sample]{}
	s := Sample{}
	_, err := s1.Next(s, map[string]any{})
	if err == nil {
		t.Errorf("must return eof")
	}
}
