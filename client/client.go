package client


import (
	"bytes"
	"fmt"
	"net"
	"time"

	"Tiny-BT-Client/bitfield"
	"Tiny-BT-Client/handshake"
	"Tiny-BT-Client/message"
	"Tiny-BT-Client/peers"
)


type Client struct {
	Conn net.Conn
	Choked bool
	Bitfield bitfield.Bitfield
	peer peers.Peer
	infoHash [20]byte
	peerID [20]byte
}


func completeHandshake(conn net.Conn, infohash, peerID [20]byte) (*handshake.Handshake, error) {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{})
	req := handshake.New(infohash, peerID)
	_, err := conn.Write(req.Serialize())
	if err != nil {
		return nil, err
	}
	res, err := handshake.Read(conn)
	if err != nil {
		return nil, err
	}
	if ! bytes.Equal(res.InfoHash[:], infohash[:]) {
		return nil, fmt.Errorf("expected infohash %x but got %x", res.InfoHash, infohash)
	}
	return res, nil
}

func recvBitfield(conn net.Conn) (bitfield.Bitfield, error) {
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	defer conn.SetDeadline(time.Time{})
	msg, err := message.Read(conn)
	if err != nil {
		return nil, err
	}
	if msg == nil {
		err := fmt.Errorf("expected bitfield but got ID: %d", msg.ID)
		return nil, err
	}
	return msg.Payload, nil
}


func New(peer peers.Peer, peerID, infohash [20]byte) (*Client, error) {
	conn, err := net.DialTimeout("tcp", peer.String(), 3*time.Second)
	if err != nil {
		return nil, err
	}
	_, err = completeHandshake(conn, infohash, peerID)
	if err != nil {
		conn.Close()
		return nil, err
	}
	bf, err := recvBitfield(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return &Client {
		Conn: conn,
		Choked: true,
		Bitfield: bf,
		peer: peer,
		infoHash: infohash,
		peerID: peerID,
	}, nil
}

// Read
func (c *Client) Read() (*message.Message, error) {
	msg, err := message.Read(c.Conn)
	return msg, err
}

// SendRequest
func (c *Client) SendRequest(index, begin, length int) error {
	req := message.FormatRequest(index, begin, length)
	_, err := c.Conn.Write(req.Serialize())
	return err
}

//SendInterested
func (c *Client) SendInterested() error {
	msg := message.Message{ID: message.MsgInterested}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// SendNotInterested
func (c *Client) SendNotInterested() error {
	msg := message.Message{ID: message.MsgNotInterested}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// SendUnchoke
func (c *Client) SendUnchoke() error {
	msg := message.Message{ID: message.MsgUnchoke}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// SendHave
func (c *Client) SendHave(index int) error {
	msg := message.FormatHave(index)
	_, err := c.Conn.Write(msg.Serialize())
	return err
}
