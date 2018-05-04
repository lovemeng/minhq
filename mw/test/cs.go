package test

import (
	"encoding/hex"
	"fmt"
	"net"

	"github.com/ekr/minq"
	"github.com/martinthomson/minhq/mw"
)

var clientAddr = &net.UDPAddr{IP: net.ParseIP("::1"), Port: 12589}
var serverAddr = &net.UDPAddr{IP: net.ParseIP("::1"), Port: 12590}

// Transport shuffles arrays of bytes from one channel to the other.
type Transport struct {
	read  <-chan []byte
	write chan<- []byte
}

// Send adds to the write side of this transport.
func (t *Transport) Send(p []byte) error {
	fmt.Printf("Transport.Send: %v\n", hex.EncodeToString(p))
	t.write <- p
	return nil
}

// Service is intended to run as a goroutine.  Service pulls from the queue of
// packets it maintains and passes those to the provided channel.
func (t *Transport) Service(addr *net.UDPAddr, c chan<- *mw.Packet) {
	for {
		p, ok := <-t.read
		if !ok {
			fmt.Printf("Transport.Service done\n")
			return
		}
		fmt.Printf("Transport.Service: %v\n", hex.EncodeToString(p))
		c <- &mw.Packet{RemoteAddr: addr, Data: p}
	}
}

// Close implements io.Closer.
func (t *Transport) Close() error {
	close(t.write)
	return nil
}

type simpleTransportFactory struct {
	t *Transport
}

func (tf *simpleTransportFactory) MakeTransport(remote *net.UDPAddr) (minq.Transport, error) {
	t := tf.t
	tf.t = nil
	return t, nil
}

// ClientServer runs a simple server that accepts a single client.
type ClientServer struct {
	ClientConnection *mw.Connection
	ServerConnection *mw.Connection
	Server           *mw.Server

	clientTransport *Transport
	serverTransport *Transport
}

// NewClientServerPair is used to support testing.
func NewClientServerPair(runServerFunc func(*minq.Server) *mw.Server,
	getServerConnectionFunc func(*mw.Server) *mw.Connection) *ClientServer {
	cs := &ClientServer{}

	a := make(chan []byte, 100)
	b := make(chan []byte, 100)
	cs.clientTransport = &Transport{a, b}
	cs.serverTransport = &Transport{b, a}

	serverConfig := minq.NewTlsConfig("localhost")
	cs.Server = runServerFunc(minq.NewServer(&simpleTransportFactory{cs.serverTransport}, &serverConfig, nil))
	go cs.serverTransport.Service(clientAddr, cs.Server.IncomingPackets)

	clientConfig := minq.NewTlsConfig("localhost")
	cs.ClientConnection = mw.NewConnection(minq.NewConnection(cs.clientTransport, minq.RoleClient, &clientConfig, nil))
	go cs.clientTransport.Service(serverAddr, cs.ClientConnection.IncomingPackets)

	clientConnected := <-cs.ClientConnection.Connected
	if cs.ClientConnection != clientConnected {
		cs.Close()
		panic("got a different client connection at the server")
	}
	
	if getServerConnectionFunc == nil {
		getServerConnectionFunc = func(s *mw.Server) *mw.Connection {
			return <-s.Connections
		}
	}

	cs.ServerConnection = getServerConnectionFunc(cs.Server)
	return cs
}

// Close releases all the resources for the pair.
func (cs *ClientServer) Close() error {
	if cs.ClientConnection != nil {
		cs.ClientConnection.Close()
	}
	if cs.ServerConnection != nil {
		cs.ServerConnection.Close()
	}
	if cs.Server != nil {
		cs.Server.Close()
	}
	if cs.serverTransport != nil {
		cs.serverTransport.Close()
	}
	if cs.clientTransport != nil {
		cs.clientTransport.Close()
	}
	return nil
}