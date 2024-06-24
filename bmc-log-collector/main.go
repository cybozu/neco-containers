package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	requestURL := "http://localhost:8750/virtualMachines"

	// URLをアクセス
	res, err := http.Get(requestURL)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	fmt.Printf("Status: %v\n", res.Status)
	fmt.Printf("StatusCode: %v\n", res.StatusCode)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))

}
