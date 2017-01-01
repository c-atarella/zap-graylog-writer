package gelf

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"strconv"
)

const (
	// VersionTag is mandatory
	VersionTag = "version"
	// HostTag is mandatory
	HostTag = "host"
	// LevelTag is mandatory
	LevelTag = "level"
)

// ZapLevelToGelfLevel maps the zap log levels to the syslog severity levels used for gelf.
// See https://en.wikipedia.org/wiki/Syslog for details.
func ZapLevelToGelfLevel(l int32) int {
	switch l {
	// DebugLevel
	case -1:
		return 7
	// InfoLevel
	case 0:
		return 6
	// WarnLevel
	case 1:
		return 4
	// ErrorLevel
	case 2:
		return 3
	// DPanicLevel
	// PanicLevel
	// FatalLevel
	case 3, 4, 5:
		return 0
	}
	// default, should never happen
	return 1
}

// Config represents the required settings for connecting the gelf data sink.
type Config struct {
	GraylogPort     int
	GraylogHostname string
	MaxChunkSize    int
}

// ZapWriteSyncer mirrors zap.WriteSyncer
type ZapWriteSyncer interface {
	io.Writer
	Sync() error
}

// New returns an implementation of ZapWriteSyncer which should be compatible with zap.WriteSyncer
func New(config Config) ZapWriteSyncer {
	return &gelf{Config: config}
}

type gelf struct {
	Config
}

func (g *gelf) Sync() error {
	// currently a noop.
	return nil
}

func (g *gelf) Write(p []byte) (int, error) {
	compressed, err := g.compress(p)
	if err != nil {
		return 0, err
	}
	chunksize := g.Config.MaxChunkSize
	length := compressed.Len()

	if length > chunksize {
		chunkCountInt := int(math.Ceil(float64(length) / float64(chunksize)))

		id := make([]byte, 8)
		rand.Read(id)

		for i, index := 0, 0; i < length; i, index = i+chunksize, index+1 {
			packet := g.createChunkedMessage(index, chunkCountInt, id, &compressed)
			_, e := g.send(packet.Bytes())
			if err != nil {
				return 0, e
			}
		}

	} else {
		_, e := g.send(compressed.Bytes())
		if err != nil {
			return 0, e
		}
	}

	fmt.Printf("Wrote data: %s\n", p)
	return len(p), nil
}

func (g *gelf) createChunkedMessage(index int, chunkCountInt int, id []byte, compressed *bytes.Buffer) bytes.Buffer {
	var packet bytes.Buffer

	chunksize := g.Config.MaxChunkSize

	packet.Write(g.intToBytes(30))
	packet.Write(g.intToBytes(15))
	packet.Write(id)

	packet.Write(g.intToBytes(index))
	packet.Write(g.intToBytes(chunkCountInt))

	packet.Write(compressed.Next(chunksize))

	return packet
}

func (g *gelf) intToBytes(i int) []byte {
	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.LittleEndian, int8(i))
	if err != nil {
		log.Printf("Uh oh! %s", err)
	}
	return buf.Bytes()
}

func (g *gelf) compress(b []byte) (bytes.Buffer, error) {
	// TODO enable compression
	var buf bytes.Buffer
	// comp := zlib.NewWriter(&buf)
	// defer comp.Close()
	// _, err := comp.Write(b)
	_, err := buf.Write(b)
	return buf, err
}

func (g *gelf) send(b []byte) (int, error) {
	var addr = g.Config.GraylogHostname + ":" + strconv.Itoa(g.Config.GraylogPort)
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Printf("Uh oh! %s", err)
		return 0, err
	}
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Printf("Uh oh! %s", err)
		return 0, err
	}
	return conn.Write(b)
}
