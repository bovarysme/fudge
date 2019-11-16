package logger

import (
	"fmt"
	"io"
	"log/syslog"
	"os"

	"bovarys.me/fudge/config"
)

func priority(name string) (syslog.Priority, error) {
	priorities := map[string]syslog.Priority{
		"emerg":   syslog.LOG_EMERG,
		"alert":   syslog.LOG_ALERT,
		"crit":    syslog.LOG_CRIT,
		"err":     syslog.LOG_ERR,
		"warning": syslog.LOG_WARNING,
		"notice":  syslog.LOG_NOTICE,
		"info":    syslog.LOG_INFO,
		"debug":   syslog.LOG_DEBUG,
	}

	priority, ok := priorities[name]
	if !ok {
		return -1, fmt.Errorf("unknown syslog priority: %q", name)
	}

	return priority, nil
}

func Writer(cfg config.LoggerConfig) (io.Writer, error) {
	switch cfg.Mode {
	case "stdout":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	case "file":
		file, err := os.OpenFile(cfg.Path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			return nil, err
		}

		return file, nil
	case "syslog":
		priority, err := priority(cfg.Priority)
		if err != nil {
			return nil, err
		}

		writer, err := syslog.New(syslog.LOG_DAEMON|priority, "fudge")
		if err != nil {
			return nil, err
		}

		return writer, nil
	default:
		return nil, fmt.Errorf("unknown logger mode: %q", cfg.Mode)
	}
}
