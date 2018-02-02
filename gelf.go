package gelf

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"log"
	"math"
	"net"
	"strconv"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// VersionTag is mandatory
	VersionTag = "version"
	// Version of supported gelf format
	Version = "1.1"
	// HostTag is mandatory
	HostTag = "host"
	// LevelTag is mandatory
	LevelTag = "level"
	// MessageKey provides the key value for gelf message field
	MessageKey = "short_message"
	TimeKey    = "timestamp"
)

func NewGelfCore(host string, appOrHostName string, fs ...zapcore.Field) (zapcore.Core, zap.Option) {
	allLevels := zap.LevelEnablerFunc(func(l zapcore.Level) bool { return true })
	syncer := New(NewDefaultConfig(host))

	config := zap.NewProductionEncoderConfig()
	config.EncodeLevel = SyslogLevelEncoder
	config.MessageKey = MessageKey
	config.TimeKey = TimeKey

	jsonEncode := zapcore.NewJSONEncoder(config)

	// see http://docs.graylog.org/en/2.3/pages/gelf.html for documentation of the gelf format
	option := zap.Fields(append(fs, zap.String(VersionTag, Version), zap.String(HostTag, appOrHostName))...)

	return zapcore.NewCore(jsonEncode, syncer, allLevels), option
}

// ZapLevelToGelfLevel maps the zap log levels to the syslog severity levels used for gelf.
// See https://en.wikipedia.org/wiki/Syslog for details.
func SyslogLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	switch l {
	case zapcore.DebugLevel:
		enc.AppendInt(7)
	case zapcore.InfoLevel:
		enc.AppendInt(6)
	case zapcore.WarnLevel:
		enc.AppendInt(4)
	case zapcore.ErrorLevel:
		enc.AppendInt(3)
	case zapcore.DPanicLevel:
		enc.AppendInt(0)
	case zapcore.PanicLevel:
		enc.AppendInt(0)
	case zapcore.FatalLevel:
		enc.AppendInt(0)
	}
}

// Config represents the required settings for connecting the gelf data sink.
type Config struct {
	GraylogPort     int
	GraylogHostname string
	MaxChunkSize    int
}

// NewDefaultConfig provides a configuration with default values for port and chunk size.
func NewDefaultConfig(host string) Config {
	return Config{GraylogPort: 12201, MaxChunkSize: 8154, GraylogHostname: host}
}

// New returns an implementation of ZapWriteSyncer which should be compatible with zap.WriteSyncer
func New(config Config) zapcore.WriteSyncer {
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

	//fmt.Printf("Wrote data: %s\n", p)
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
