package node

import (
	"context"

	"github.com/arcology/3rd-party/tm/common"
	"github.com/arcology/Monaco/core/config"
	"github.com/arcology/Monaco/core/p2p"
	"github.com/arcology/Monaco/core/version"
	"github.com/arcology/Monaco/tmlibs/log"
	"github.com/arcology/component-lib/actor"
	"github.com/arcology/component-lib/kafka/lib"
	"github.com/arcology/p2p-svc/mempool"
)

type Node struct {
	Logger log.Logger

	sw       *p2p.Switch
	config   *config.Config
	uploader *lib.ComOutgoing
}

func NewNode(ctx context.Context, config *config.Config, kafkaAddrs []string, logger log.Logger) *Node {
	uploader := &lib.ComOutgoing{}
	uploader.Start(kafkaAddrs, map[string]string{actor.MsgTxRemotes: "remote-txs"}, "uploader")
	mempoolReactor := mempool.NewMempoolReactor(logger.With("module", "mempool"), uploader)

	sw := p2p.NewSwitch(config.P2P)
	sw.SetLogger(logger.With("module", "p2p"))
	sw.AddReactor("MEMPOOL", mempoolReactor)

	return &Node{
		Logger:   logger,
		sw:       sw,
		config:   config,
		uploader: uploader,
	}
}

func (n *Node) Start() error {
	protocol, address := common.ProtocolAndAddress(n.config.P2P.ListenAddress)
	l := p2p.NewDefaultListener(protocol, address, n.config.P2P.SkipUPNP, n.Logger.With("module", "p2p"))
	n.sw.AddListener(l)

	// Generate node PrivKey
	// TODO: pass in like privValidator
	nodeKey, err := p2p.LoadOrGenNodeKey(n.config.NodeKeyFile())
	if err != nil {
		return err
	}
	n.Logger.Info("P2P Node ID", "ID", nodeKey.ID(), "file", n.config.NodeKeyFile())

	nodeInfo := n.makeNodeInfo(nodeKey.ID())
	n.sw.SetNodeInfo(nodeInfo)
	n.sw.SetNodeKey(nodeKey)

	// Start the switch (the P2P server).
	err = n.sw.Start()
	if err != nil {
		return err
	}

	// Always connect to persistent peers
	if n.config.P2P.PersistentPeers != "" {
		err = n.sw.DialPeersAsync(nil, common.SplitAndTrim(n.config.P2P.PersistentPeers, ",", " "), true)
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *Node) Stop() {
	n.uploader.Stop()
	n.sw.Stop()
}

func (n *Node) makeNodeInfo(nodeID p2p.ID) p2p.NodeInfo {
	nodeInfo := p2p.NodeInfo{
		ID:      nodeID,
		Network: "p2p",
		Version: version.Version,
		Channels: []byte{
			mempool.MempoolChannel,
		},
		Moniker: n.config.Moniker,
	}

	if !n.sw.IsListening() {
		return nodeInfo
	}

	p2pListener := n.sw.Listeners()[0]
	p2pHost := p2pListener.ExternalAddress().IP.String()
	p2pPort := p2pListener.ExternalAddress().Port
	nodeInfo.ListenAddr = common.Fmt("%v:%v", p2pHost, p2pPort)

	return nodeInfo
}
