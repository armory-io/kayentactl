package logger

import (
	"bytes"

	log "github.com/sirupsen/logrus"
)

type PlainLogger struct{}

func (p *PlainLogger) Format(entry *log.Entry) ([]byte, error) {
	var b bytes.Buffer
	b.WriteString(entry.Message)
	b.WriteByte('\n')
	return b.Bytes(), nil
}
