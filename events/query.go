package events

import (
	"fmt"

	tmQuery "github.com/tendermint/tendermint/libs/pubsub/query"
)

func QueryForTxExecution(txHash []byte) *tmQuery.Query {
	return tmQuery.MustParse(fmt.Sprintf("gallactic.events.tx.hash='%X'", txHash))
}
