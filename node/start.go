package node

import (
	"context"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"

	"github.com/arcology-network/3rd-party/tm/cli"
	tmCommon "github.com/arcology-network/3rd-party/tm/common"
	cfg "github.com/arcology-network/Monaco/core/config"
	"github.com/arcology-network/Monaco/tmlibs/log"
	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/component-lib/actor"
	"github.com/arcology-network/component-lib/kafka/lib"
	clog "github.com/arcology-network/component-lib/log"
	"github.com/arcology-network/p2p-svc/mempool"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start p2p service Daemon",
	RunE:  startCmd,
}

var (
// config        = cfg.DefaultConfig()
// kafkaAddrsStr = ""
// isdebug       = false
// logcfg        = ""
// nidx          = 0
// nname         = ""
)

func init() {
	flags := StartCmd.Flags()

	flags.String("mqaddr", "localhost:9092", "host:port of kafka ")

	flags.Bool("debug", false, "debug mode")

	flags.String("logcfg", "./log.toml", "log conf path")

	flags.Int("nidx", 0, "node index in cluster")
	flags.String("nname", "node1", "node name in cluster")
	flags.Int64("sendrate", 10240000, "p2p send rate in byte")
	flags.Int64("recvrate", 10240000, "p2p receive rate in byte")
	flags.String("persistent_peers", "", "Comma-delimited ID@host:port persistent peers")
	flags.String("laddr", "tcp://0.0.0.0:36656", "Node listen address. (0.0.0.0:0 means any interface, any port)")
	flags.String("node_key", "node_key.json", "Node key file")
}

func startCmd(cmd *cobra.Command, args []string) error {

	config := cfg.DefaultConfig()
	config.P2P.PersistentPeers = viper.GetString("persistent_peers")
	config.P2P.ListenAddress = viper.GetString("laddr")
	config.BaseConfig.NodeKey = viper.GetString("node_key")
	config.P2P.SendRate = viper.GetInt64("sendrate")
	config.P2P.RecvRate = viper.GetInt64("recvrate")

	kafkaAddrsStr := viper.GetString("mqaddr")

	if len(config.P2P.PersistentPeers) == 0 {
		panic("persistent_peers not set.")
	}

	logname := "p2p.log"
	rootDir := viper.GetString(cli.HomeFlag)
	//create logger
	if err := tmCommon.EnsureDir(path.Join(rootDir, "log"), 0777); err != nil {
		tmCommon.PanicSanity(err.Error())
	}
	logfile, err := os.OpenFile(path.Join(rootDir, "log", logname), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		tmCommon.PanicSanity(err.Error())
	}

	clog.InitLog("p2p_com.log", viper.GetString("logcfg"), "p2p", viper.GetString("nname"), viper.GetInt("nidx"))

	logger := log.NewTMLogger(log.NewSyncWriter(logfile))
	logger = log.NewFilter(logger, log.AllowInfo())

	if viper.GetBool("debug") {
		logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	}
	//return logger
	logger = logger.With("svc", "p2p")

	n := NewNode(context.Background(), config, strings.Split(kafkaAddrsStr, ","), logger)
	n.Start()

	downloader := &lib.ComIncoming{}
	downloader.Start(
		strings.Split(kafkaAddrsStr, ","),
		[]string{"local-txs"},
		[]string{"txLocals"},
		"p2p", "downloader",
		func(msg *actor.Message) error {
			//logger.Info("Received new message", msg)
			msg.Name = actor.MsgTxRemotes
			msgBytes, err := common.GobEncode(msg)
			if err != nil {
				logger.Info("GobEncode failed with error", err)
				return err
			}
			n.sw.Broadcast(mempool.MempoolChannel, msgBytes)
			return nil
		},
	)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	<-sigterm

	n.Stop()
	downloader.Stop()
	return nil
}
