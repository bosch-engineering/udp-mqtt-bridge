package pingpong

import (
	"fmt"
	"time"
)

// PingPong represents the ping-pong mechanism.
type PingPong struct {
	pingChan chan struct{}
	pongChan chan struct{}
}

// NewPingPong creates a new PingPong instance.
func NewPingPong() *PingPong {
	return &PingPong{
		pingChan: make(chan struct{}),
		pongChan: make(chan struct{}),
	}
}

// Start begins the ping-pong process.
func (pp *PingPong) Start() {
	go pp.ping()
	go pp.pong()

	// Initiate the ping-pong process
	pp.pingChan <- struct{}{}
}

// ping handles the ping logic.
func (pp *PingPong) ping() {
	for {
		<-pp.pingChan
		fmt.Println("Ping")
		time.Sleep(1 * time.Second)
		pp.pongChan <- struct{}{}
	}
}

// pong handles the pong logic.
func (pp *PingPong) pong() {
	for {
		<-pp.pongChan
		fmt.Println("Pong")
		time.Sleep(1 * time.Second)
		pp.pingChan <- struct{}{}
	}
}
