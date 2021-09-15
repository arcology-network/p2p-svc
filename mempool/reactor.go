package mempool

import (
	"fmt"

	"github.com/arcology-network/Monaco/core/p2p"
	"github.com/arcology-network/Monaco/tmlibs/log"
	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/component-lib/actor"
	"github.com/arcology-network/component-lib/kafka/lib"
)

const (
	MempoolChannel = byte(0x30)

	maxMsgSize = 1048576 // 1MB
)

type MempoolReactor struct {
	p2p.BaseReactor

	logger   log.Logger
	uploader *lib.ComOutgoing
}

func NewMempoolReactor(logger log.Logger, uploader *lib.ComOutgoing) *MempoolReactor {
	reactor := &MempoolReactor{
		logger:   logger,
		uploader: uploader,
	}
	reactor.BaseReactor = *p2p.NewBaseReactor("MempoolReactor", reactor)
	return reactor
}

func (reactor *MempoolReactor) GetChannels() []*p2p.ChannelDescriptor {
	return []*p2p.ChannelDescriptor{
		{
			ID:                MempoolChannel,
			Priority:          5,
			SendQueueCapacity: 1024,
		},
	}
}

func (reactor *MempoolReactor) AddPeer(peer p2p.Peer) {

}

func (reactor *MempoolReactor) RemovePeer(peer p2p.Peer, reason interface{}) {

}

func (reactor *MempoolReactor) Receive(chID byte, src p2p.Peer, msgBytes []byte) {
	var msg actor.Message
	err := common.GobDecode(msgBytes, &msg)
	if err != nil {
		reactor.logger.Error("Error decoding message", "src", src, "chId", chID, "msg", msg, "err", err, "bytes", msgBytes)
		reactor.Switch.StopPeerForError(src, err)
		return
	}
	reactor.logger.Debug("Receive", "src", src, "chId", chID, "msg", msg)

	switch msg.Name {
	case actor.MsgTxRemotes:
		reactor.uploader.Send(&msg)
	default:
		reactor.logger.Error(fmt.Sprintf("Unknown message type %v", msg.Name))
	}
}
