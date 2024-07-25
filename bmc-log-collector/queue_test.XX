package main

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"time"
)

/*
Tests machineListReader(), which reads a CSV file with a specified path and sets it into a structure.
*/
var _ = Describe("Machines Queue v2", Ordered, func() {

	var que MessageQueue

	BeforeAll(func() {
		fmt.Println("Get Machines List")
		que.queue = make(chan Machine, 1000)
	})

	Context("Manipulate queue", func() {
		It("Put Queue", func() {
			var qTemp []Machine
			m1 := Machine{
				Serial: "ABC123",
				BmcIP:  "192.168.0.1",
				NodeIP: "172.16.0.1",
			}
			qTemp = append(qTemp, m1)
			m2 := Machine{
				Serial: "DEF123",
				BmcIP:  "192.168.0.2",
				NodeIP: "172.16.0.2",
			}
			qTemp = append(qTemp, m2)

			que.put3(qTemp)
			Expect(len(que.queue)).To(Equal(2))

		})

		It("Get que length, expect = 2", func(ctx SpecContext) {
			fmt.Println(que.queue)
			m := que.len2()
			Expect(m).To(Equal(2))
		})

		It("Put Queue again", func() {
			var qTemp []Machine

			m1 := Machine{
				Serial: "GHI123",
				BmcIP:  "192.168.0.3",
				NodeIP: "172.16.0.3",
			}
			qTemp = append(qTemp, m1)

			m2 := Machine{
				Serial: "JKLM123",
				BmcIP:  "192.168.0.4",
				NodeIP: "172.16.0.4",
			}
			qTemp = append(qTemp, m2)
			que.put3(qTemp)

			Expect(que.len2()).To(Equal(4))
		})

		It("Get que length, expect = 4", func(ctx SpecContext) {
			m := que.len2()
			Expect(m).To(Equal(4))
		})

		It("Get machine1 from que", func(ctx SpecContext) {
			done := make(chan interface{})
			go func() {
				defer GinkgoRecover()
				m := que.get2()
				Expect(m.Serial).To(Equal("ABC123"))
				Expect(m.BmcIP).To(Equal("192.168.0.1"))
				Expect(m.NodeIP).To(Equal("172.16.0.1"))
				close(done)
			}()
			Eventually(done).Should(BeClosed())
		}, SpecTimeout(3*time.Second))

		It("Get machine2 from que", func(ctx SpecContext) {
			done := make(chan interface{})
			go func() {
				defer GinkgoRecover()
				m := que.get2()
				Expect(m.Serial).To(Equal("DEF123"))
				Expect(m.BmcIP).To(Equal("192.168.0.2"))
				Expect(m.NodeIP).To(Equal("172.16.0.2"))
				close(done)
			}()
			Eventually(done).Should(BeClosed())
		}, SpecTimeout(time.Second))

		It("Get que length, expect = 2", func(ctx SpecContext) {
			m := que.len2()
			Expect(m).To(Equal(2))
		})
	})
})
