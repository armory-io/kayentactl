package logger

import (
	"bytes"

	"github.com/fatih/color"

	log "github.com/sirupsen/logrus"
)

type ColorizedLogger struct{}

func (p *ColorizedLogger) Format(entry *log.Entry) ([]byte, error) {
	var b bytes.Buffer
	switch entry.Level {
	case log.ErrorLevel, log.FatalLevel, log.PanicLevel:
		b.WriteString(color.RedString(entry.Message))
	case log.WarnLevel:
		b.WriteString(color.YellowString(entry.Message))
	default:
		b.WriteString(entry.Message)
	}
	b.WriteByte('\n')
	return b.Bytes(), nil
}

type PlainLogger struct{}

func (p *PlainLogger) Format(entry *log.Entry) ([]byte, error) {
	var b bytes.Buffer
	b.WriteString(entry.Message)
	b.WriteByte('\n')
	return b.Bytes(), nil
}
