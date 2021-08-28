package main

import (
	"encoding/gob"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/arcology/component-lib/actor"
	"github.com/arcology/component-lib/kafka/lib"
	"github.com/arcology/component-lib/log"
)

func main() {
	log.InitLog("feeder.log", "log.toml", "feeder", "feeder", 0)

	kafka := os.Args[1]
	n, _ := strconv.Atoi(os.Args[2])
	xlen, _ := strconv.Atoi(os.Args[3])
	ylen, _ := strconv.Atoi(os.Args[4])

	uploader := &lib.ComOutgoing{}
	uploader.Start([]string{kafka}, map[string]string{"txLocals": "local-txs"}, "feeder")
	gob.Register([][]byte{})

	data := make([][]byte, xlen)
	for i := 0; i < xlen; i++ {
		data[i] = make([]byte, ylen)
	}

	start := time.Now()
	fmt.Println("START TIME: ", start)
	for i := 0; i < n; i++ {
		uploader.Send(&actor.Message{
			Name: "txLocals",
			Data: data,
		})
	}
	fmt.Println(time.Since(start))

	time.Sleep(5 * time.Second)
}
