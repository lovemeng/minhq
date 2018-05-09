package minhq

import (
	"errors"

	"github.com/ekr/minq"
	"github.com/martinthomson/minhq/hc"
	"github.com/martinthomson/minhq/mw"
)

// ErrStreamBlocked is used to indicate that there are no streams available.
// TODO: consider blocking until a stream is available.
var ErrStreamBlocked = errors.New("Unable to open a new stream for the request")

// ClientConnection is a connection specialized for use by clients.
type ClientConnection struct {
	connection
}

// NewClientConnection wraps an instance of minq.Connection.
func NewClientConnection(mwc *mw.Connection, config *Config) *ClientConnection {
	hq := &ClientConnection{
		connection: connection{
			Connection: *mwc,
			decoder:    hc.NewQcramDecoder(config.DecoderTableCapacity),
			encoder:    hc.NewQcramEncoder(0, 0),
		},
	}
	hq.Init(hq)
	return hq
}

// HandleFrame is for dealing with those frames that Connection can't.
func (c *ClientConnection) HandleFrame(t FrameType, f byte, r FrameReader) error {
	return ErrInvalidFrame
}

// Fetch makes a request.
func (c *ClientConnection) Fetch(method string, target string, headers ...hc.HeaderField) (*ClientRequest, error) {
	<-c.Connected
	if c.GetState() != minq.StateEstablished {
		return nil, errors.New("connection not open")
	}

	allHeaders, err := buildRequestHeaderFields(method, target, headers)
	if err != nil {
		return nil, err
	}

	requestID := c.nextRequestID()
	s := newStream(c.CreateStream())
	if s == nil {
		return nil, ErrStreamBlocked
	}
	_, err = s.WriteVarint(requestID.id)
	if err != nil {
		return nil, err
	}

	err = writeHeaderBlock(c.encoder, c.headersStream, s, requestID, allHeaders)
	if err != nil {
		return nil, err
	}

	responseChannel := make(chan *ClientResponse)
	req := &ClientRequest{
		Response:        responseChannel,
		OutgoingMessage: newOutgoingMessage(&c.connection, s, requestID, allHeaders),
	}

	go req.readResponse(s, c, responseChannel)
	return req, nil
}
