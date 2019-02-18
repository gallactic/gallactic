package logger

import (
	log "github.com/inconshreveable/log15"
	tmLog "github.com/tendermint/tendermint/libs/log"
)

type tendermintLogger struct {
	log.Logger
}

func NewLogger() *tendermintLogger {
	return &tendermintLogger{}
}

func (tml *tendermintLogger) With(keyvals ...interface{}) tmLog.Logger {
	return &tendermintLogger{
		log.New(keyvals...),
	}
}
