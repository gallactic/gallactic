package burrow

import (
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/exec"
)

type noopEventSink struct{}

func (*noopEventSink) Call(call *exec.CallEvent, exception *errors.Exception) {}
func (*noopEventSink) Log(log *exec.LogEvent)                                 {}
