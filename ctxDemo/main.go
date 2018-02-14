package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	// Pass a context with a timeout to tell a blocking function that it
	// should abandon its work after the timeout elapses.
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	ticker := time.Tick(1 * time.Second)
	timeout := time.After(3 * time.Second)
	for {
		select {
		case <-ticker:
			fmt.Println("Ticking")

		case <-timeout:
			fmt.Println("Timeout")
			cancel()

		case <-ctx.Done():
			fmt.Println(ctx.Err()) // prints "context deadline exceeded"
			return
		}
	}

}
