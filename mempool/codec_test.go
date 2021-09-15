package mempool

import (
	"testing"

	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/component-lib/actor"
)

func TestCodec(t *testing.T) {
	msg := actor.Message{
		Name: actor.MsgTxRemotes,
		Data: [][]byte{{1, 2, 3}, {4, 5, 6}},
	}

	msgBytes, err := common.GobEncode(&msg)
	if err != nil {
		t.Error("fail")
		return
	}

	var msg2 actor.Message
	err = common.GobDecode(msgBytes, &msg2)
	if err != nil {
		t.Error("fail")
		return
	}

	if len(msg2.Data.([][]byte)) != 2 {
		t.Error("fail")
	}
}
