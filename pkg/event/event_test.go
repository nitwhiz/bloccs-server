package event

import (
	"testing"
)

type TestEventSource struct {
	Id    string
	Extra int
}

func (s *TestEventSource) GetId() string {
	return s.Id
}

func (s *TestEventSource) RLock() {}

func (s *TestEventSource) RUnlock() {}

func TestEvent_New(t *testing.T) {
	src := TestEventSource{Id: "1337", Extra: 42}

	evnt := New("testType", &src, nil)

	if evnt.Source != &src {
		t.Fatalf("source is not the same")
	}

	if (evnt.Source).(*TestEventSource).Extra != 42 {
		t.Fatalf("incorrect event source data")
	}
}
