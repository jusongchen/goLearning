package main

import (
	"fmt"
	"time"

	"context"
	"os"
	"os/signal"
)

type server struct {
}

func (s *server) stop() {
	fmt.Println("Stopping Server...")
	fmt.Println("Release resources ...")
	fmt.Println("Done cleaning up...")
}

func (s *server) serve(ctx context.Context) {

LoopFor:
	for {
		select {
		case <-ctx.Done(): //either timeout or ask to cancel
			s.stop() //cleanup
			break LoopFor
		default:
		}
		//do the work
		fmt.Printf("Working on task, press Ctrl+C to stop ...\n")
		time.Sleep(time.Millisecond * 600)
	}

}

func main() {
	s := server{}

	const TIMEOUT_SECONDS = 8
	//Server will timeout in N seconds
	ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT_SECONDS*time.Second)

	//handling Ctrl+C
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for _ = range signalChan {
			fmt.Println("\nReceived an interrupt, stopping services...")
			cancel() //request server to return ASAP
		}
	}()

	s.serve(ctx)
}
