package main

import (
	"fmt"
	"os"
)

func main() {
	if err := actor(nil); err != nil {
		fmt.Printf("Error %+v", err)
		os.Exit(1)
	}
	return
}
