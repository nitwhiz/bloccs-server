package event

import "testing"

type TestEventPayload struct {
	Position int
}

func (s *TestEventPayload) RLock() {}

func (s *TestEventPayload) RUnlock() {}

func TestBus_GenericWithoutPayload(t *testing.T) {
	b := NewBus()

	callCount1 := 0
	callCount2 := 0

	expectedCallCount1 := 3
	expectedCallCount2 := 2

	b.Subscribe("testEvent1", func(event *Event) {
		callCount1 += 1
	})

	b.Subscribe("testEvent2", func(event *Event) {
		callCount2 += 1
	})

	b.Start()

	e1 := New("testEvent1", &TestEventSource{Id: "1337"}, nil)
	e2 := New("testEvent1", &TestEventSource{Id: "1337"}, nil)
	e3 := New("testEvent2", &TestEventSource{Id: "1337"}, nil)
	e4 := New("testEvent1", &TestEventSource{Id: "1337"}, nil)
	e5 := New("testEvent2", &TestEventSource{Id: "1337"}, nil)

	b.Publish(e1)
	b.Publish(e2)
	b.Publish(e3)
	b.Publish(e4)
	b.Publish(e5)

	b.Stop()

	if callCount1 != expectedCallCount1 {
		t.Fatalf("callCount1 is %d instead of %d", callCount1, expectedCallCount1)
	}

	if callCount2 != expectedCallCount2 {
		t.Fatalf("callCount2 is %d instead of %d", callCount2, expectedCallCount2)
	}
}

func TestBus_GenericWithPayload(t *testing.T) {
	b := NewBus()

	var recievedPayloads []int

	callCount1 := 0
	callCount2 := 0

	expectedCallCount1 := 3
	expectedCallCount2 := 2

	b.Subscribe("testEvent1", func(event *Event) {
		callCount1 += 1
		recievedPayloads = append(recievedPayloads, event.Payload.(*TestEventPayload).Position)
	})

	b.Subscribe("testEvent2", func(event *Event) {
		callCount2 += 1
		recievedPayloads = append(recievedPayloads, event.Payload.(*TestEventPayload).Position)
	})

	b.Start()

	e1 := New("testEvent1", &TestEventSource{Id: "1337"}, &TestEventPayload{
		Position: 0,
	})
	e2 := New("testEvent1", &TestEventSource{Id: "1337"}, &TestEventPayload{
		Position: 1,
	})
	e3 := New("testEvent2", &TestEventSource{Id: "1337"}, &TestEventPayload{
		Position: 2,
	})
	e4 := New("testEvent1", &TestEventSource{Id: "1337"}, &TestEventPayload{
		Position: 3,
	})
	e5 := New("testEvent2", &TestEventSource{Id: "1337"}, &TestEventPayload{
		Position: 4,
	})

	b.Publish(e1)
	b.Publish(e2)
	b.Publish(e3)
	b.Publish(e4)
	b.Publish(e5)

	b.Stop()

	if callCount1 != expectedCallCount1 {
		t.Fatalf("callCount1 is %d instead of %d", callCount1, expectedCallCount1)
	}

	if callCount2 != expectedCallCount2 {
		t.Fatalf("callCount2 is %d instead of %d", callCount2, expectedCallCount2)
	}

	for i, p := range recievedPayloads {
		if p != i {
			t.Fatalf("event order is mixed up")
		}
	}
}
