package main

/*
  Read the machine list and access iDRAC mock.
  Verify anti-duplicate filter.
*/
import (
	"context"
	"fmt"
	"testing"
)

func TestUpdateBMCLogCollectorConfigMap(t *testing.T) {

	//var m1,m2 Machine
	var ml []Machine
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

	c := client{}
	err := c.updateBMCLogCollectorConfigMap(context.Background(), ml)
	fmt.Println("err", err)

}
