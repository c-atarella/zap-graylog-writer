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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ZapLevelToGelfLevel(tt.args.l); got != tt.want {
				t.Errorf("ZapLevelToGelfLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}
