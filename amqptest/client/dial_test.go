package client

import (
	"errors"
	"testing"
	"time"

	"github.com/tiago4orion/wabbit"
	"github.com/tiago4orion/wabbit/amqptest/server"
)

var rabbitmqPort = "35672"

// WaitOK blocks until rabbitmq can accept connections on
// <ctn ip address>:5672
func waitRabbitOK(amqpuri string) error {
	var err error
	var counter = 0
dial:
	if counter > 120 {
		panic("Impossible to connect to rabbitmq")
	}

	_, err = Dial(amqpuri)
	if err != nil {
		time.Sleep(500 * time.Millisecond)
		counter++
		goto dial
	}

	return nil
}

// TestDial test a simple connection to rabbitmq.
// If the rabbitmq disconnects will not be tested here!
func TestDial(t *testing.T) {
	amqpuri := "amqp://guest:guest@localhost:35672/%2f"

	// Should fail
	conn, err := Dial(amqpuri)

	if err == nil {
		t.Error("No backend started... Should fail")
		return
	}

	server := server.NewServer(amqpuri)
	server.Start()

	err = waitRabbitOK(amqpuri)

	if err != nil {
		t.Error(err)
		return
	}

	conn, err = Dial(amqpuri)

	if err != nil || conn == nil {
		t.Error(err)
		return
	}

	server.Stop()
}

func TestAutoRedial(t *testing.T) {
	var err error

	amqpuri := "amqp://guest:guest@localhost:35672/%2f"

	server := server.NewServer(amqpuri)
	server.Start()
	defer server.Stop()

	err = waitRabbitOK(amqpuri)

	if err != nil {
		t.Error(err)
		return
	}

	conn, err := Dial(amqpuri)

	if err != nil {
		t.Error(err)
		return
	}

	defer conn.Close()
	redialErrors := make(chan wabbit.Error)

	done := make(chan bool)
	conn.AutoRedial(redialErrors, done)

	// required goroutine to consume connection error messages
	go func() {
		for {
			// discards the connection errors
			<-redialErrors
		}
	}()

	server.Stop()

	// concurrently starts the rabbitmq after 1 second
	go func() {
		time.Sleep(10 * time.Second)
		server.Start()
	}()

	select {
	case <-time.After(20 * time.Second):
		err = errors.New("Timeout exceeded. AMQP reconnect failed")
	case <-done:
		err = nil
	}

	if err != nil {
		t.Errorf("Client doesn't reconnect in 3 seconds: %s", err.Error())
		return
	}
}

func TestChannelMock(t *testing.T) {
	var channel wabbit.Channel

	// rabbitmq.Channel satisfies wabbit.Channel interface
	channel = new(Channel)

	if channel == nil {
		t.Error("Maybe wabbit.Channel interface does not mock amqp.Channel correctly")
	}
}

func TestConnMock(t *testing.T) {
	var conn wabbit.Conn

	// rabbitmq.Conn satisfies wabbit.Conn interface
	conn = &Conn{}

	if conn == nil {
		t.Error("Maybe wabbit.Conn interface does not mock amqp.Conn correctly")
	}
}