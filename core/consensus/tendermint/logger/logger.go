package logger

import (
	"github.com/hyperledger/burrow/logging"
	tmLog "github.com/tendermint/tendermint/libs/log"
)

type tendermintLogger struct {
	logger *logging.Logger
}

func NewLogger(logger *logging.Logger) *tendermintLogger {
	return &tendermintLogger{
		logger: logger,
	}
}

func (tml *tendermintLogger) Info(msg string, keyvals ...interface{}) {
	tml.logger.InfoMsg(msg, keyvals...)
}

func (tml *tendermintLogger) Error(msg string, keyvals ...interface{}) {
	tml.logger.InfoMsg(msg, keyvals...)
}

func (tml *tendermintLogger) Debug(msg string, keyvals ...interface{}) {
	tml.logger.TraceMsg(msg, keyvals...)
}

func (tml *tendermintLogger) With(keyvals ...interface{}) tmLog.Logger {
	return &tendermintLogger{
		logger: tml.logger.With(keyvals...),
	}
}
