package main

/*
マシンリストを読んで、iDRACへアクセスする。
重複防止機能を確認する
*/

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	//. "github.com/onsi/gomega"
)

var _ = Describe("Collecting iDRAC Logs", Ordered, func() {

	// Start iDRAC Simulator
	BeforeAll(func() {
		fmt.Println("Start iDRAC Simulator IT-2")
		start_iDRAC_Simulator_it2_idrac1()
		start_iDRAC_Simulator_it2_idrac2()
		time.Sleep(10 * time.Second)
	})

	BeforeEach(func() {
		os.Setenv("BMC_USER", "user")
		os.Setenv("BMC_PASS", "pass")
		os.Setenv("LOG_COLLECTOR", "1")
	})

	Context("IT main-loop test", func() {
		It("normal running", func(ctx SpecContext) {
			GinkgoWriter.Println("start")
			os.Setenv("LOG_COLLECTOR", "2")
			global_ctx := context.Background()
			global_ctx, global_cancel := context.WithCancel(global_ctx)
			global_wg.Add(1)
			go runMainLoop(global_ctx)
			defer global_cancel()
			time.Sleep(30 * time.Second)
			global_cancel()
			global_wg.Done()
			global_wg.Wait()
			//time.Sleep(20 * time.Second)
			//Expect(err).NotTo(HaveOccurred())
		}, SpecTimeout(60*time.Second))
		//})
		/*
			It("normal running", func(ctx SpecContext) {
				GinkgoWriter.Println("start")
				os.Setenv("LOG_COLLECTOR", "1")
				global_ctx := context.Background()
				global_ctx, global_cancel := context.WithCancel(global_ctx)
				global_wg.Add(1)
				go runMainLoop(global_ctx)
				defer global_cancel()
				time.Sleep(30 * time.Second)
				global_cancel()
				global_wg.Wait()
				//time.Sleep(20 * time.Second)
				//Expect(err).NotTo(HaveOccurred())
			}, SpecTimeout(60*time.Second))
			//})
		*/
	})

	AfterAll(func() {
		fmt.Println("Shutdown webserver for IT-2")
		time.Sleep(5 * time.Second)
		err := exec.Command("rm", "pointers/*").Run()
		fmt.Println(err)
	})

})
