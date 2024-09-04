package utils

import (
	"fmt"
	"net"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/gookit/slog"
)

// UDP connection struct
type UDPConn struct {
	conn        *net.UDPConn
	receiveChan chan []byte
}

// Method to receive UDP packets
func (u *UDPConn) Receive() <-chan []byte {
	return u.receiveChan
}

// NewConnection creates a new UDP connection.
func NewConnection(ip string, portIn int) (*UDPConn, error) {
	slog.Infof("Initializing UDP connection on port %d", portIn)

	// Initialize UDP connection
	addr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(ip, fmt.Sprint(portIn)))
	if err != nil {
		slog.Errorf("Error resolving UDP address: %v", err)
		return nil, err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		slog.Errorf("Error listening on UDP port: %v", err)
		return nil, err
	}

	udpConn := &UDPConn{
		conn:        conn,
		receiveChan: make(chan []byte),
	}

	go udpConn.listen()

	return udpConn, nil
}

// listen method to continuously read from the UDP connection
func (u *UDPConn) listen() {
	defer close(u.receiveChan)
	buf := make([]byte, 1024)
	for {
		n, _, err := u.conn.ReadFromUDP(buf)
		if err != nil {
			slog.Infof("Error receiving UDP packet: %v", err)
			continue
		}
		// Copy the received data to avoid overwriting
		data := make([]byte, n)
		copy(data, buf[:n])
		u.receiveChan <- data
	}
}

// Method to send UDP packets
func (u *UDPConn) Send(ip string, port int, ce cloudevents.Event) error {
	// Convert the CloudEvent to a string
	slog.Infof("Sending UDP CloudEvent message: %s %s", ce.ID(), ce.Type())
	message, err := MarshallCloudEvent(ce)
	if err != nil {
		slog.Errorf("Error marshalling CloudEvent: %v", err)
		return err
	}

	addr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(ip, fmt.Sprint(port)))
	if err != nil {
		return err
	}

	_, err = u.conn.WriteToUDP(message, addr)
	if err != nil {
		slog.Infof("Error sending UDP packet: %v", err)
	}
	return err
}

// Connection represents a UDP connection.
type Connection struct {
	conn *net.UDPConn
}

// RemoteAddress resolves a UDP address from a given IP and port.
func (c *Connection) RemoteAddress(ip string, port int) (*net.UDPAddr, error) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		slog.Errorf("Error resolving UDP address: %v", err)
		return nil, err
	}
	return addr, nil
}

func (c *Connection) Receive() chan []byte {
	ch := make(chan []byte)
	go func() {
		defer close(ch)
		for {
			buf := make([]byte, 1024) // Move buffer initialization inside the loop
			n, _, err := c.conn.ReadFromUDP(buf)
			if err != nil {
				slog.Infof("Error receiving UDP packet: %v", err) // Log the error instead of fatal
				continue                                          // Continue listening for packets
			}
			ch <- buf[:n]
		}
	}()
	return ch
}

func (c *Connection) Send(ip string, port int, ce cloudevents.Event) error {
	// Convert the CloudEvent to a string
	slog.Infof("Sending UDP CloudEvent message: %s %s", ce.ID(), ce.Type())
	message, err := MarshallCloudEvent(ce)
	if err != nil {
		slog.Errorf("Error marshalling CloudEvent: %v", err)
		return err
	}

	addr, err := c.RemoteAddress(ip, port)
	if err != nil {
		return err
	}

	_, err = c.conn.WriteToUDP(message, addr)
	if err != nil {
		slog.Infof("Error sending UDP packet: %v", err) // Log the error
	}
	return err
}
