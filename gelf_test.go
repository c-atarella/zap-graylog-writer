package gelf

import "testing"

func _TestWriteIntegration(t *testing.T) {
	conf := Config{GraylogHostname: "127.0.0.1", GraylogPort: 12201, MaxChunkSize: 8154}

	sut := New(conf)
	msg := "{\"version\": \"1.1\", \"host\": \"test\", \"short_message\": \"hello world from zapp brannigan\"}"

	n, err := sut.Write([]byte(msg))
	if n != len(msg) {
		t.Errorf("Expected for n value %v but got value %v", len(msg), n)
	}
	if err != nil {
		t.Errorf("Expected no error but got %v", err)
	}
}

func TestNewDefaultConfig(t *testing.T) {
	result := NewDefaultConfig("foo")

	if result.GraylogHostname != "foo" {
		t.Errorf("Expected for result.GraylogHostname value %v but got value %v", "foo", result.GraylogHostname)
	}
	if result.GraylogPort != 12201 {
		t.Errorf("Expected for result.GraylogPort value %v but got value %v", 8154, result.GraylogPort)
	}
	if result.MaxChunkSize != 8154 {
		t.Errorf("Expected for result.MaxChunkSize value %v but got value %v", 8154, result.MaxChunkSize)
	}
}
