package logger

import (
	log "github.com/inconshreveable/log15"
	tmFlags "github.com/tendermint/tendermint/libs/cli/flags"
	tmLog "github.com/tendermint/tendermint/libs/log"
)

type tendermintLogger struct {
	log.Logger
}

func NewLoggerF(filter string, keyvals ...interface{}) tmLog.Logger {
	l := &tendermintLogger{
		log.New(),
	}

	tLogger, err := tmFlags.ParseLogLevel(filter, l, "*:info")
	if err != nil {
		panic("Unable to start tendermint logger: " + err.Error())
	}

	return tLogger
}

func NewLogger(keyvals ...interface{}) tmLog.Logger {
	return NewLoggerF("*:debug")
}

func (tml *tendermintLogger) With(keyvals ...interface{}) tmLog.Logger {
	return NewLogger(keyvals)
}
