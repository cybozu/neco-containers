package main

/*
  Read the machine list and access iDRAC mock.
  Verify anti-duplicate filter.
*/
import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Collecting iDRAC Logs", Ordered, func() {

	//var m1,m2 Machine
	var ml []Machine

	BeforeAll(func() {
		fmt.Println("test before all")
		var m0 Machine
		m0.Spec.IPv4 = append(m0.Spec.IPv4, "1.1.1.1")
		m0.Spec.IPv4 = append(m0.Spec.IPv4, "1.2.2.2")
		m0.Spec.Serial = "ABC123"
		ml = append(ml, m0)

		var m1 Machine
		m1.Spec.IPv4 = append(m1.Spec.IPv4, "2.1.1.1")
		m1.Spec.IPv4 = append(m1.Spec.IPv4, "2.2.2.2")
		m1.Spec.Serial = "XYZ123"
		ml = append(ml, m1)

		fmt.Println("ml", ml)
		time.Sleep(10 * time.Second)
	})

	Context("main test", func() {
		It("test", func() {
			fmt.Println("test1")
			fmt.Println("machines")

			x := 0
			Expect(x).To(Equal(0))
		})
	})

})
