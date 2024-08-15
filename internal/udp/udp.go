package udp

import (
	"fmt"
	"log"
	"net"
)

// UDP connection struct
type UDPConn struct {
	receiveChan chan []byte
}

// Method to receive UDP packets
func (u *UDPConn) Receive() <-chan []byte {
	return u.receiveChan
}

// NewConnection creates a new UDP connection.
func NewConnection(ip string, portIn, portOut int) (*Connection, error) {
	log.Printf("Initializing UDP connection on port %d", portIn)
	log.Printf("Sending UDP packets to port %d", portOut)

	// Initialize UDP connection
	addr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(ip, fmt.Sprint(portIn)))
	if err != nil {
		log.Fatalf("Error resolving UDP address: %v", err)
		return nil, err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("Error initializing UDP connection: %v", err)
		return nil, err
	}

	return &Connection{conn: conn}, nil
}

// Connection represents a UDP connection.
type Connection struct {
	conn *net.UDPConn
}

// RemoteAddress resolves a UDP address from a given IP and port.
func (c *Connection) RemoteAddress(ip string, port int) (*net.UDPAddr, error) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		log.Printf("Error resolving UDP address: %v", err)
		return nil, err
	}
	return addr, nil
}

func (c *Connection) Receive() chan []byte {
	ch := make(chan []byte)
	go func() {
		defer close(ch)
		buf := make([]byte, 1024)
		for {
			n, _, err := c.conn.ReadFromUDP(buf)
			if err != nil {
				log.Printf("Error receiving UDP packet: %v", err)
				return
			}
			ch <- buf[:n]
		}
	}()
	return ch
}

func (c *Connection) Send(data []byte, ip string, port int) error {
	addr, err := c.RemoteAddress(ip, port)
	if err != nil {
		return err
	}

	log.Printf("Sending UDP packet: %s to %s:%d", string(data), ip, port)

	_, err = c.conn.WriteToUDP(data, addr)
	if err != nil {
		log.Printf("Error sending UDP packet: %v", err)
		return err
	}

	log.Printf("Sent UDP packet: %s", string(data))

	return nil
}
