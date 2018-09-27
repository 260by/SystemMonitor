package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
	"github.com/260by/tools/gconfig"
	"github.com/260by/SystemMonitor/data"
	gssh "github.com/260by/SystemMonitor/ssh"
)

const (
	version         = "0.1"
)

type Config struct {
	Interval int
	Database struct {
		Driver string
		Dsn string
		ShowSQL bool
		Migrate bool
	}
}

// 命令行处理
func usage() {
	fmt.Fprintf(os.Stderr, `monitor Version %s
Usage: monitor options

Options:
`, version)
	flag.PrintDefaults()
}

func main() {
	var configFile = flag.String("config", "config.toml", "Configration file")
	// var interval = flag.Int("interval", 30, "Get data interval time(second)")
	// var migrate = flag.Bool("migrate", false, "Sync database table structure")
	flag.Usage = usage

	flag.Parse()
	log.SetPrefix("monitor: ")
	log.SetFlags(0)

	var config = Config{}
	err := gconfig.Parse(*configFile, &config)
	if err != nil {
		panic(err)
	}

	if config.Database.Migrate {
		orm, err := data.Connect(config.Database.Driver, config.Database.Dsn, config.Database.ShowSQL)
		if err != nil {
			panic(err)
		}
		err = data.Migrate(orm)
		if err != nil {
			panic(err)
		}
		if err == nil {
			fmt.Println("Sync database table structure is success.")
			os.Exit(0)
		}
	}

	for {
		t1 := time.Now()

		monitors := make([]data.Monitor, 0)
		var wg sync.WaitGroup

		// 连接数据库
		orm, err := data.Connect(config.Database.Driver, config.Database.Dsn, config.Database.ShowSQL)
		if err != nil {
			panic(err)
		}
		// defer db.Close()

		assetsList := []data.Assets{}
		orm.Find(&assetsList)
		// db.Find(&assetsList)

		for _, assets := range assetsList {
			wg.Add(1)
			go func(userName, ip, authenticate string, port int) {
				client, err := gssh.Connect(userName, ip, port, authenticate)
				if err != nil {
					log.Fatalln(err)
				}
	
				monitor := data.GetMonitorData(client)
				monitors = append(monitors, monitor)
				defer wg.Done()
			}(assets.UserName, assets.IP, assets.Authenticate, assets.Port)
		}
		wg.Wait()
		// fmt.Println(monitors)
		orm.Insert(&monitors)
		orm.Close()

		fmt.Println("Get data success.")
		fmt.Println(time.Since(t1))
		time.Sleep(time.Second * time.Duration(config.Interval))
	}

}
