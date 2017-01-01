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

func TestZapLevelToGelfLevel(t *testing.T) {
	type args struct {
		l int32
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "invalid level", args: args{l: -2}, want: 1},
		{name: "debug level", args: args{l: -1}, want: 7},
		{name: "info level", args: args{l: 0}, want: 6},
		{name: "warn level", args: args{l: 1}, want: 4},
		{name: "error level", args: args{l: 2}, want: 3},
		{name: "dpanic level", args: args{l: 3}, want: 0},
		{name: "panic level", args: args{l: 4}, want: 0},
		{name: "fatal level", args: args{l: 5}, want: 0},
		{name: "invalid level", args: args{l: 6}, want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ZapLevelToGelfLevel(tt.args.l); got != tt.want {
				t.Errorf("ZapLevelToGelfLevel() = %v, want %v", got, tt.want)
			}
		})
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
