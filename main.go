package main

import (
	"flag"
	"fmt"
	"github.com/260by/SystemMonitor/data"
	gssh "github.com/260by/SystemMonitor/ssh"
	"log"
	"os"
	"sync"
	"time"
)

const (
	version         = "0.1"
)

// 命令行处理
func usage() {
	fmt.Fprintf(os.Stderr, `monitor Version %s
Usage: monitor options

Options:
`, version)
	flag.PrintDefaults()
}

func main() {
	var interval = flag.Int("interval", 30, "Get data interval time(second)")
	flag.Usage = usage

	flag.Parse()
	log.SetPrefix("monitor: ")
	log.SetFlags(0)

	// 连接数据库
	db, err := data.Connect()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	assetsList := []data.Assets{}
	db.Find(&assetsList)

	for {
		t1 := time.Now()
		var wg sync.WaitGroup
		for _, assets := range assetsList {
			wg.Add(1)
			go func(userName, ip, privateKey string, port int) {
				client, err := gssh.Connect(userName, ip, port, privateKey)
				if err != nil {
					log.Fatalln(err)
				}
	
				data.InsertData(client, db)
				defer wg.Done()
			}(assets.UserName, assets.IP, assets.PrivateKey, assets.Port)
		}
		wg.Wait()

		fmt.Println("Get data success.")
		fmt.Println(time.Since(t1))
		time.Sleep(time.Second * time.Duration(*interval))
	}

}
