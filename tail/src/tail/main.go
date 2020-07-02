package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/well"
	"github.com/nxadm/tail"
)

func main() {
	follow := flag.Bool("f", false, "follow the growth of the file")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("usage: ./tail [-f] [FILE]...")
		os.Exit(1)
	}

	for _, file := range flag.Args() {
		file := file
		well.Go(func(ctx context.Context) error {
			t, err := tail.TailFile(file, tail.Config{Follow: *follow})
			if err != nil {
				return err
			}
			go func() {
				<-ctx.Done()
				t.Stop()
				t.Cleanup()
			}()
			for line := range t.Lines {
				fmt.Println(line.Text)
			}
			return nil
		})
	}
	well.Stop()
	err := well.Wait()
	if err != nil && !well.IsSignaled(err) {
		log.ErrorExit(err)
	}
}
