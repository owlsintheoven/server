package models

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"owlsintheoven/learning-go/common"
	"strconv"
)

// constant
const (
	DSTPORT_BYTES = 2
	DSTIP_BYTES   = 4

	REQUEST_GRANTED            = "request granted"
	REQUEST_REJECTED_OR_FAILED = "request rejected or failed"

	CONNECT string = "CONNECT"
	BIND           = "BIND"
)

type Socks4Req struct {
	VN     string
	CMD    string
	UserID string
	Port   int
	IP     net.IP
}

func ParseSocks4Request(reader *bufio.Reader) (*Socks4Req, error) {
	version, err := common.ReadByte(reader)
	if err != nil {
		return nil, fmt.Errorf("reading VN error: %w", err)
	}
	if version != 0x04 {
		return nil, fmt.Errorf("invalid VN")
	}

	cmd, err := common.ReadByte(reader)
	if err != nil {
		return nil, fmt.Errorf("reading CMD error: %w", err)
	}

	var mode string
	if cmd == 0x01 {
		mode = CONNECT
	} else if cmd != 0x02 {
		mode = BIND
	} else {
		return nil, fmt.Errorf("invalid CMD")
	}

	dstPort, err := common.ReadBytes(reader, DSTPORT_BYTES)
	if err != nil {
		return nil, fmt.Errorf("reading DstPort error: %w", err)
	}

	dstIP, err := common.ReadBytes(reader, DSTIP_BYTES)
	if err != nil {
		return nil, fmt.Errorf("reading DstIP error: %w", err)
	}

	userID, err := common.ReadBytesTilNil(reader)
	if err != nil {
		return nil, fmt.Errorf("reading UserID error: %w", err)
	}

	return &Socks4Req{
		VN:     "4",
		CMD:    mode,
		UserID: string(userID),
		Port:   int(binary.BigEndian.Uint16(dstPort[:])),
		IP:     dstIP,
	}, nil
}

func (s *Socks4Req) GetIPString() string {
	return s.IP.String()
}

func (s *Socks4Req) GetPortString() string {
	return strconv.Itoa(s.Port)
}

func (s *Socks4Req) IsConnect() bool {
	return s.CMD == CONNECT
}

func FormResponse(result string) []byte {
	response := make([]byte, 8)
	response[0] = 0x00
	switch result {
	case REQUEST_GRANTED:
		response[1] = 0x5a
	case REQUEST_REJECTED_OR_FAILED:
		response[1] = 0x5b
	}
	return response
}
