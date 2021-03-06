package boltdb

import (
	"encoding/gob"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/yehohanan7/flux/cqrs"
)

const DB_PATH = "flux_test.db"

type EventPayload struct {
	Data string
}

var _ = Describe("Bolt Event Store", func() {

	var store EventStore

	BeforeEach(func() {
		gob.Register(EventPayload{})
		store = NewBoltStore(DB_PATH)
	})

	AfterEach(func() {
		os.Remove(DB_PATH)
	})

	It("Should save events", func() {
		expected := NewEvent("sample_aggregate", 0, EventPayload{"payload"})

		err := store.SaveEvents("aggregate-1", []Event{expected})

		Expect(err).To(BeNil())
		Expect(store.GetEvent(expected.Id)).To(Equal(expected))
	})

	It("Should save event metadata", func() {
		e1 := NewEvent("sample_aggregate", 0, EventPayload{"payload"})
		e2 := NewEvent("sample_aggregate", 1, EventPayload{"payload"})
		e3 := NewEvent("sample_aggregate", 2, EventPayload{"payload"})

		err := store.SaveEvents("aggregate-1", []Event{e1, e2, e3})

		Expect(err).To(BeNil())
		Expect(store.GetEventMetaDataFrom(0, 1)).To(HaveLen(1))
		Expect(store.GetEventMetaDataFrom(0, 2)).To(HaveLen(2))
		Expect(store.GetEventMetaDataFrom(0, 3)).To(HaveLen(3))
		Expect(store.GetEventMetaDataFrom(0, 10)).To(HaveLen(3))
	})

	It("Should retrieve event meta data with all attributes", func() {
		event := NewEvent("sample_aggregate", 0, EventPayload{"payload"})

		err := store.SaveEvents("aggregate-1", []Event{event})

		Expect(err).To(BeNil())
		meta := store.GetEventMetaDataFrom(0, 1)[0]
		Expect(meta).To(Equal(event.EventMetaData))
	})

	It("Should retreive events by aggregate Id", func() {
		expected := make([]Event, 20)
		for i := 0; i < 20; i++ {
			expected[i] = NewEvent("sample_aggregate", i, EventPayload{"payload"})
		}
		store.SaveEvents("aggregate-1", expected)

		events := store.GetEvents("aggregate-1")

		Expect(events).To(HaveLen(20))
		for i, e := range events {
			Expect(e).To(Equal(expected[i]))
		}
	})

	It("Should handle empty events", func() {
		events := store.GetEvents("unknown")

		Expect(events).To(HaveLen(0))
	})

	It("Should reject aggregate with existing version", func() {
		e1 := NewEvent("sample_aggregate", 0, EventPayload{"payload"})
		e2 := NewEvent("sample_aggregate", 1, EventPayload{"payload"})
		err := store.SaveEvents("aggregate-1", []Event{e1, e2})
		Expect(err).To(BeNil())

		err = store.SaveEvents("aggregate-1", []Event{e1})
		Expect(err).ShouldNot(BeNil())
		Expect(store.GetEvents("aggregate-1")).To(Equal([]Event{e1, e2}))
	})

	It("Should reject invalid events", func() {
		e1 := NewEvent("sample_aggregate", 0, EventPayload{"payload"})
		e2 := NewEvent("sample_aggregate", 1, EventPayload{"payload"})
		store.SaveEvents("aggregate-1", []Event{e1, e2})

		for _, e := range []Event{NewEvent("sample_aggregate", 3, EventPayload{"payload"}), NewEvent("sample_aggregate", 1, EventPayload{"payload"})} {
			Expect(store.SaveEvents("aggregate-1", []Event{e})).ShouldNot(BeNil())
		}

		Expect(store.SaveEvents("aggregate-1", []Event{NewEvent("sample_aggregate", 2, EventPayload{"payload"})})).Should(BeNil())
	})

	It("Should retrieve event meta data from specific offset", func() {
		events := []Event{}
		for i := 0; i < 15; i++ {
			events = append(events, NewEvent("sample_aggregate", i, EventPayload{"payload"}))
		}

		err := store.SaveEvents("aggregate-1", events)
		Expect(err).To(BeNil())
		Expect(len(store.GetEventMetaDataFrom(0, 1))).To(Equal(1))
		Expect(len(store.GetEventMetaDataFrom(0, 15))).To(Equal(15))
		Expect(len(store.GetEventMetaDataFrom(5, 15))).To(Equal(11))
	})

})
