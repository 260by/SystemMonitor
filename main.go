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
	version         = "1.0"
	CPUDefaultRefresh = 3 // 以秒为单位的默认刷新CPU间隔
)

var (
	interval   int
)

func init() {
	flag.IntVar(&interval, "interval", 30, "Get data interval time(second)")
	flag.Usage = usage
}

// 命令行处理
func usage() {
	fmt.Fprintf(os.Stderr, `rtop Version %s
Usage: rtop options

Options:
`, version)
	flag.PrintDefaults()
}

func main() {
	// t1 := time.Now()

	flag.Parse()
	log.SetPrefix("top: ")
	log.SetFlags(0)

	// interval = conf.Interval
	// if interval == 0 {
	// 	interval = DEFAULT_REFRESH
	// }

	// 连接数据库
	db, err := data.Conn()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	assetsList := []data.Assets{}
	db.Find(&assetsList)

	for {
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
		time.Sleep(time.Second * time.Duration(interval))
	}

	// fmt.Println(time.Since(t1))
}
