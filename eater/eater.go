package main

import (
	"encoding/gob"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/arcology-network/component-lib/actor"
	"github.com/arcology-network/component-lib/kafka/lib"
	"github.com/arcology-network/component-lib/log"
)

func main() {
	log.InitLog("eater.log", "log.toml", "eater", "eater", 0)

	kafka := os.Args[1]
	n, _ := strconv.Atoi(os.Args[2])

	count := 0
	isFirst := true
	var start time.Time

	gob.Register([][]byte{})
	downloader := &lib.ComIncoming{}
	downloader.Start(
		[]string{kafka},
		[]string{"remote-txs"},
		[]string{"txRemotes"},
		"eater", "downloader",
		func(msg *actor.Message) error {
			if isFirst {
				start = time.Now()
				isFirst = false
			}
			count++
			if count == n {
				fmt.Println(time.Since(start))
				fmt.Println("END TIME: ", time.Now())
				isFirst = true
				count = 0
			}
			return nil
		},
	)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	<-sigterm
}
