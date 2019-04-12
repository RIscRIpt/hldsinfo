package hldsinfo

import (
	"errors"
	"io"
	"net"
	"time"
)

// ErrUnexpectedResponse is returned when server responds with unparseable response.
var ErrUnexpectedResponse = errors.New("unexpected response from the server")

// InfoExtra contains extra information about the server, if Info.EDF is != 0.
type InfoExtra struct {
	Port         uint16 `json:"port,omitempty"`
	SteamID      uint64 `json:"steamid,omitempty"`
	SourceTVPort uint16 `json:"sourcetv_port,omitempty"`
	SourceTVName string `json:"sourcetv_name,omitempty"`
	Keywords     string `json:"keywords,omitempty"`
	GameID       uint64 `json:"game_id,omitempty"`
}

// Info contains information about the server
type Info struct {
	Header      byte      `json:"-"`
	Protocol    byte      `json:"protocol"`
	Name        string    `json:"name"`
	Map         string    `json:"map"`
	Folder      string    `json:"folder"`
	Game        string    `json:"game"`
	ID          uint16    `json:"id"`
	Players     byte      `json:"players"`
	MaxPlayers  byte      `json:"max_players"`
	Bots        byte      `json:"bots"`
	ServerType  string    `json:"server_type"`
	Environment string    `json:"environment"`
	Visibility  byte      `json:"visibility"`
	VAC         byte      `json:"vac"`
	Version     string    `json:"version"`
	EDF         byte      `json:"-"`
	ExtraData   InfoExtra `json:"extra_data,omitempty"`
}

var a2sInfo = []byte("\xFF\xFF\xFF\xFFTSource Engine Query\x00")

type buffer struct {
	data   []byte
	offset int
}

func writeAllDeadline(conn net.Conn, data []byte, deadline time.Time) error {
	conn.SetWriteDeadline(deadline)
	for i := 0; i < len(a2sInfo); {
		n, err := conn.Write(data[i:])
		if err != nil {
			return err
		}
		i += n
	}
	return nil
}

func (b *buffer) readByte() (byte, error) {
	if b.offset >= len(b.data) {
		return 0, io.EOF
	}
	result := b.data[b.offset]
	b.offset++
	return result, nil
}

func (b *buffer) readChar() (string, error) {
	if b.offset >= len(b.data) {
		return "", io.EOF
	}
	result := string(b.data[b.offset])
	b.offset++
	return result, nil
}

func (b *buffer) readUInt16() (uint16, error) {
	var err error
	var bytes [2]byte
	for i := 0; i < len(bytes); i++ {
		bytes[i], err = b.readByte()
		if err != nil {
			return 0, err
		}
	}
	return (uint16(bytes[0]) << 0) |
		(uint16(bytes[1]) << 8), nil
}
func (b *buffer) readUInt32() (uint32, error) {
	var err error
	var bytes [4]byte
	for i := 0; i < len(bytes); i++ {
		bytes[i], err = b.readByte()
		if err != nil {
			return 0, err
		}
	}
	return (uint32(bytes[0]) << 0) |
		(uint32(bytes[1]) << 8) |
		(uint32(bytes[2]) << 16) |
		(uint32(bytes[3]) << 24), nil
}

func (b *buffer) readUInt64() (uint64, error) {
	var err error
	var bytes [8]byte
	for i := 0; i < len(bytes); i++ {
		bytes[i], err = b.readByte()
		if err != nil {
			return 0, err
		}
	}
	return (uint64(bytes[0]) << 0) |
		(uint64(bytes[1]) << 8) |
		(uint64(bytes[2]) << 16) |
		(uint64(bytes[3]) << 24) |
		(uint64(bytes[4]) << 32) |
		(uint64(bytes[5]) << 40) |
		(uint64(bytes[6]) << 48) |
		(uint64(bytes[7]) << 56), nil
}

func (b *buffer) readString() (string, error) {
	var bytes []byte
	for {
		b, err := b.readByte()
		if err != nil {
			return "", err
		}
		if b == 0 {
			break
		}
		bytes = append(bytes, b)
	}
	return string(bytes), nil
}

// Get returns Info about the server with specified address.
func Get(address string, deadline time.Time) (*Info, error) {
	var err error
	var conn net.Conn
	if deadline.IsZero() {
		conn, err = net.Dial("udp4", address)
	} else {
		conn, err = net.DialTimeout("udp4", address, deadline.Sub(time.Now()))
	}
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	err = writeAllDeadline(conn, a2sInfo, deadline)
	if err != nil {
		return nil, err
	}

	info := new(Info)
	var buf buffer
	buf.data = make([]byte, 1024)

	conn.SetReadDeadline(deadline)
	n, err := conn.Read(buf.data)
	if err != nil {
		return nil, err
	}
	buf.data = buf.data[:n]

	var header uint32
	header, err = buf.readUInt32()
	if err != nil {
		return nil, err
	}
	if header != ^uint32(0) {
		return nil, ErrUnexpectedResponse
	}

	info.Header, err = buf.readByte()
	if err != nil {
		return nil, err
	}
	if info.Header != 0x49 {
		return nil, ErrUnexpectedResponse
	}

	info.Protocol, err = buf.readByte()
	if err != nil {
		return nil, err
	}

	info.Name, err = buf.readString()
	if err != nil {
		return nil, err
	}

	info.Map, err = buf.readString()
	if err != nil {
		return nil, err
	}

	info.Folder, err = buf.readString()
	if err != nil {
		return nil, err
	}

	info.Game, err = buf.readString()
	if err != nil {
		return nil, err
	}

	info.ID, err = buf.readUInt16()
	if err != nil {
		return nil, err
	}

	info.Players, err = buf.readByte()
	if err != nil {
		return nil, err
	}

	info.MaxPlayers, err = buf.readByte()
	if err != nil {
		return nil, err
	}

	info.Bots, err = buf.readByte()
	if err != nil {
		return nil, err
	}

	info.ServerType, err = buf.readChar()
	if err != nil {
		return nil, err
	}

	info.Environment, err = buf.readChar()
	if err != nil {
		return nil, err
	}

	info.Visibility, err = buf.readByte()
	if err != nil {
		return nil, err
	}

	info.VAC, err = buf.readByte()
	if err != nil {
		return nil, err
	}

	info.Version, err = buf.readString()
	if err != nil {
		return nil, err
	}

	info.EDF, err = buf.readByte()
	if err != nil {
		return nil, err
	}

	if info.EDF&0x80 != 0 {
		info.ExtraData.Port, err = buf.readUInt16()
		if err != nil {
			return nil, err
		}
	}

	if info.EDF&0x10 != 0 {
		info.ExtraData.SteamID, err = buf.readUInt64()
		if err != nil {
			return nil, err
		}
	}

	if info.EDF&0x40 != 0 {
		info.ExtraData.SourceTVPort, err = buf.readUInt16()
		if err != nil {
			return nil, err
		}
		info.ExtraData.SourceTVName, err = buf.readString()
		if err != nil {
			return nil, err
		}
	}

	if info.EDF&0x20 != 0 {
		info.ExtraData.Keywords, err = buf.readString()
		if err != nil {
			return nil, err
		}
	}

	if info.EDF&0x01 != 0 {
		info.ExtraData.GameID, err = buf.readUInt64()
		if err != nil {
			return nil, err
		}
	}

	return info, nil
}
