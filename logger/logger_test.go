package logger

import (
	"io"
	"log/syslog"
	"os"
	"testing"

	"bovarys.me/fudge/config"
)

func TestPriority(t *testing.T) {
	tests := []struct {
		name     string
		priority syslog.Priority
	}{
		{"emerg", syslog.LOG_EMERG},
		{"wtf", -1},
	}

	for _, test := range tests {
		p, _ := priority(test.name)
		if p != test.priority {
			t.Errorf("wrong priority from name %q: got %v want %v",
				test.name, p, test.priority)
		}
	}
}

func TestWriter(t *testing.T) {
	tests := []struct {
		cfg    config.LoggerConfig
		writer io.Writer
	}{
		{config.LoggerConfig{Mode: "stdout"}, os.Stdout},
		{config.LoggerConfig{Mode: "stderr"}, os.Stderr},
		{config.LoggerConfig{Mode: "null"}, nil},
	}

	for _, test := range tests {
		writer, _ := Writer(test.cfg)
		if writer != test.writer {
			t.Errorf("wrong writer from mode %q: got %v want %v",
				test.cfg.Mode, writer, test.writer)
		}
	}
}
