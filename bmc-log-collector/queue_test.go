package main

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sync"
	"time"
)

/*
Tests machineListReader(), which reads a CSV file with a specified path and sets it into a structure.
*/
var _ = Describe("Machines Queue", Ordered, func() {

	//var w sync.WaitGroup
	var m sync.Mutex
	var q []Machine
	var que Queue

	BeforeAll(func() {
		fmt.Println("Get Machines List")
	})

	Context("Manipulate queue", func() {
		It("Put Queue", func() {
			que = Queue{
				queue: q,
				mu:    &m,
			}
			m1 := Machine{
				Serial: "ABC123",
				BmcIP:  "192.168.0.1",
				NodeIP: "172.16.0.1",
			}
			m2 := Machine{
				Serial: "DEF123",
				BmcIP:  "192.168.0.2",
				NodeIP: "172.16.0.2",
			}
			q = append(q, m1)
			q = append(q, m2)
			que.put(q)
			Expect(len(que.queue)).To(Equal(2))
		})

		It("Get que length, expect = 2", func(ctx SpecContext) {
			m := que.len()
			Expect(m).To(Equal(2))
		})

		It("Get machine1 from que", func(ctx SpecContext) {
			done := make(chan interface{})
			go func() {
				defer GinkgoRecover()
				m := que.get()
				Expect(m.Serial).To(Equal("ABC123"))
				Expect(m.BmcIP).To(Equal("192.168.0.1"))
				Expect(m.NodeIP).To(Equal("172.16.0.1"))
				close(done)
			}()
			Eventually(done).Should(BeClosed())
		}, SpecTimeout(time.Second))

		It("Get machine2 from que", func(ctx SpecContext) {
			done := make(chan interface{})
			go func() {
				defer GinkgoRecover()
				m := que.get()
				Expect(m.Serial).To(Equal("DEF123"))
				Expect(m.BmcIP).To(Equal("192.168.0.2"))
				Expect(m.NodeIP).To(Equal("172.16.0.2"))
				close(done)
			}()
			Eventually(done).Should(BeClosed())
		}, SpecTimeout(time.Second))

		It("Get que length, expect = 0", func(ctx SpecContext) {
			m := que.len()
			Expect(m).To(Equal(0))
		})

	})
})
