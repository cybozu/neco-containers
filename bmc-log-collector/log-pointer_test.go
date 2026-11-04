package main

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("pointer store", func() {
	It("keeps pointers and removes the ones of disappeared machines", func() {
		store := newPointerStore()
		Expect(store.get("ABCDEF")).To(Equal(LastPointer{}))

		store.put("ABCDEF", LastPointer{LastReadId: 1, SelRead: true})
		store.put("WITHDRAWED", LastPointer{LastReadId: 2, SelRead: true})
		Expect(store.get("ABCDEF")).To(Equal(LastPointer{LastReadId: 1, SelRead: true}))

		store.deleteDisappeared([]Machine{{Serial: "ABCDEF", BmcIP: "10.0.0.1", NodeIP: "10.1.0.1"}})
		Expect(store.get("ABCDEF")).To(Equal(LastPointer{LastReadId: 1, SelRead: true}))
		Expect(store.get("WITHDRAWED")).To(Equal(LastPointer{}))
	})
})

var _ = Describe("startup lookback", func() {
	It("treats entries created within the lookback as recent", func() {
		c := selCollector{startupLookback: time.Hour}
		requestTime, err := time.Parse(time.RFC3339, "2024-01-01T12:00:00Z")
		Expect(err).NotTo(HaveOccurred())

		Expect(c.isRecent(SystemEventLog{Create: "2024-01-01T11:30:00Z"}, requestTime)).To(BeTrue())
		Expect(c.isRecent(SystemEventLog{Create: "2024-01-01T11:00:00Z"}, requestTime)).To(BeTrue())
		Expect(c.isRecent(SystemEventLog{Create: "2024-01-01T05:59:59-05:00"}, requestTime)).To(BeFalse())
		Expect(c.isRecent(SystemEventLog{Create: "broken"}, requestTime)).To(BeTrue())
	})
})
