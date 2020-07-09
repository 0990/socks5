package logconfig

import (
	"fmt"
	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
	"io"
)

func InitLogrus(name string, maxMB int, level logrus.Level) {
	formatter := &logrus.TextFormatter{
		DisableColors:    true,
		DisableTimestamp: false,
		TimestampFormat:  "2006-01-02 15:04:05",
	}
	logrus.SetFormatter(formatter)
	logrus.SetLevel(level)
	logrus.AddHook(NewDefaultHook(name, maxMB))
}

type DefaultHook struct {
	writer    io.Writer
	errWriter io.Writer
	fmt       logrus.Formatter
}

func NewDefaultHook(name string, maxSize int) *DefaultHook {
	formatter := &logrus.TextFormatter{
		DisableColors:    true,
		DisableTimestamp: false,
	}

	writer := &lumberjack.Logger{
		Filename:   fmt.Sprintf("%s.log", name),
		MaxSize:    maxSize,
		MaxAge:     100,
		MaxBackups: 100,
		LocalTime:  true,
		Compress:   false,
	}

	errWriter := &lumberjack.Logger{
		Filename:   fmt.Sprintf("%s_err.log", name),
		MaxSize:    maxSize,
		MaxAge:     100,
		MaxBackups: 100,
		LocalTime:  true,
		Compress:   false,
	}

	return &DefaultHook{
		writer:    writer,
		errWriter: errWriter,
		fmt:       formatter,
	}
}

func (p *DefaultHook) Fire(entry *logrus.Entry) error {
	data, err := p.fmt.Format(entry)
	if err != nil {
		return err
	}
	if entry.Level < logrus.ErrorLevel {
		_, err = p.writer.Write(data)
	} else {
		_, err = p.errWriter.Write(data)
	}
	return err
}

func (p *DefaultHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
