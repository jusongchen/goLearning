package main

import (
	"fmt"
	"log"
	"time"

	"database/sql"

	"context"
	"os"
	"os/signal"

	"github.com/pkg/errors"

	_ "gopkg.in/rana/ora.v4"
)

type server struct {
	db *sql.DB
}

func (s *server) stop() {
	fmt.Println("Stopping Server...")
	if s.db != nil {
		fmt.Println("Close DB connection")
		s.db.Close()
	}
	fmt.Println("Done cleaning up...")
}

//openDB: open Oracle database
func (s *server) openDB(oraUser, oraPasswd, oraConn string) error {

	fmt.Println("Try making connection to DB ...")
	connStr := fmt.Sprintf("%s/%s@%s", oraUser, oraPasswd, oraConn)
	db, err := sql.Open("ora", connStr)
	// db, err := sql.Open("oci8", connStr)

	if err != nil {
		return errors.Wrapf(err, "connect to oracle %s as user %s failed", oraConn, oraUser)
	}

	//make a SQL query call to make sure the DB server works
	var n int
	err = db.QueryRow("select 1 from dual").Scan(&n)
	if err != nil {
		return errors.Wrapf(err, "connect to oracle %s as user %s failed", oraConn, oraUser)
	}

	if n != 1 {
		panic(fmt.Sprintf("OpenDB:`select 1 from dual` fail. Expecting 1 get %d", n))
	}

	s.db = db
	return nil
}

func (s *server) serve(ctx context.Context) {

	var (
		oraUser   = "test"
		oraPasswd = "test_pw"
		oraConn   = "//localhost/orcl"
	)

	sleepDuration := time.Second * 4
	fmt.Printf("Sleep for %v ..., press CTRL+C to quit", sleepDuration)

	select {
	case <-ctx.Done(): //either timeout or ask to cancel
		return
	case <-time.After(sleepDuration):
	}

	if err := s.openDB(oraUser, oraPasswd, oraConn); err != nil {
		log.Printf("%v", err)
	}

LoopFor:
	for {
		select {
		case <-ctx.Done(): //either timeout or ask to cancel
			s.stop() //cleanup
			break LoopFor
		default:
		}
		time.Sleep(time.Millisecond * 400)
		fmt.Printf("Working on task, press Ctrl+C to stop ...\n")
	}

}

func main() {

	s := server{}

	//Server will timeout in 10 seconds
	timeoutDuration := time.Second * 8
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)

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
