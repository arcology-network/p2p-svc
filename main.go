package main

import (
	"os"

	"github.com/arcology-network/3rd-party/tm/cli"
	"github.com/arcology-network/p2p-svc/node"
	"github.com/spf13/cobra"
)

func main() {

	// http.Handle("/metrics", promhttp.Handler())
	// go http.ListenAndServe(":19010", nil)

	var p2pCmd = &cobra.Command{
		Use:   "p2p-svc",
		Short: " p2p service",
		Long:  ` p2p service,It's the p2p logical structure as a conductor`,
	}

	p2pCmd.AddCommand(
		node.StartCmd,
	)

	cmd := cli.PrepareMainCmd(p2pCmd, "BC", os.ExpandEnv("$HOME/monacos/p2p"))
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
