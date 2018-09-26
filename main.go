package main

import (
	"strconv"
	"flag"
	"fmt"
	"github.com/260by/SystemMonitor/data"
	"github.com/260by/SystemMonitor/sys"
	"github.com/jinzhu/gorm"
	myssh "github.com/260by/SystemMonitor/ssh"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"sort"
	"strings"
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
				client, err := myssh.Connect(userName, ip, port, privateKey)
				if err != nil {
					log.Fatalln(err)
				}
	
				getStats(client, db)
				wg.Done()
			}(assets.UserName, assets.IP, assets.PrivateKey, assets.Port)
		}
		wg.Wait()
		fmt.Println("Get data success.")
		time.Sleep(time.Second * time.Duration(interval))
	}

	// var wg sync.WaitGroup
	// for _, assets := range assetsList {
	// 	wg.Add(1)
	// 	go func(userName, ip, privateKey string, port int) {
	// 		client, err := myssh.Connect(userName, ip, port, privateKey)
	// 		if err != nil {
	// 			log.Fatalln(err)
	// 		}

	// 		getStats(client, db)
	// 		wg.Done()
	// 	}(assets.UserName, assets.IP, assets.PrivateKey, assets.Port)
	// }
	// wg.Wait()

	// fmt.Println(time.Since(t1))
}

func getStats(client *ssh.Client, db *gorm.DB) {
	montior := data.Monitor{}

	stats := sys.Stats{}
	sys.GetAllStats(client, &stats)

	var preCPU sys.CPURaw
	sys.GetCPU(client, &stats, preCPU, CPUDefaultRefresh)

	memUsed := stats.MemTotal - stats.MemFree - stats.MemBuffers - stats.MemCached
	memUsePercent := float64(memUsed) / float64(stats.MemTotal)

	ip := strings.Split(client.RemoteAddr().String(), ":")[0]

	var diskUse string
	for _, fs := range stats.FSInfos {
		use := float32(fs.Used) / float32(fs.Used+fs.Free)
		diskUse += fmt.Sprintf("%s:%v ", fs.MountPoint, use)
	}

	montior.CreateTime = time.Now().Unix()
	montior.HostName = stats.Hostname
	montior.CPUUse = stats.CPU.User
	montior.MemUse = memUsePercent
	montior.DiskUse = diskUse
	montior.Load1, _ = strconv.ParseFloat(stats.Load1, 64)
	montior.Load5, _ = strconv.ParseFloat(stats.Load5, 64)
	montior.Load10, _ = strconv.ParseFloat(stats.Load10, 64)
	montior.IP = ip

	db.Create(&montior)
}

func showStats(client *ssh.Client, interval int) {
	stats := sys.Stats{}
	sys.GetAllStats(client, &stats)

	var preCPU sys.CPURaw
	sys.GetCPU(client, &stats, preCPU, interval)

	var ip []string
	if len(stats.NetIntf) > 0 {
		keys := make([]string, 0, len(stats.NetIntf))
		for intf := range stats.NetIntf {
			keys = append(keys, intf)
		}
		sort.Strings(keys)
		for _, intf := range keys {
			info := stats.NetIntf[intf]
			for _, i := range info.IPv4 {
				ipv4 := strings.Split(i, "/")[0]
				ip = append(ip, ipv4)
			}
		}
	}

	memUsed := stats.MemTotal - stats.MemFree - stats.MemBuffers - stats.MemCached
	memUsePercent := float64(memUsed) / float64(stats.MemTotal) * 100

	fmt.Printf(`HostName: %s    IP: %s

CPU Used: %.2f%%

Memory: Used: %.2f%%
	Total	= %s
	Used	= %s
	Buffers	= %s
	Cached	= %s
	Free	= %s`,
		stats.Hostname, strings.Join(ip, ","), stats.CPU.User,
		memUsePercent,
		fmtBytes(stats.MemTotal), fmtBytes(memUsed),
		fmtBytes(stats.MemBuffers), fmtBytes(stats.MemCached),
		fmtBytes(stats.MemFree))

	fmt.Println()

	if len(stats.FSInfos) > 0 {
		fmt.Println("Filesystems:")
		for _, fs := range stats.FSInfos {
			use := float32(fs.Used) / float32(fs.Used+fs.Free) * 100
			fmt.Printf(`
	%s: Use %.2f%%
	    Total = %s
	    Used  = %s
	    Free  = %s`,
				fs.MountPoint, use,
				fmtBytes(fs.Used+fs.Free),
				fmtBytes(fs.Used),
				fmtBytes(fs.Free),
			)
		}
	}
	fmt.Println()
	fmt.Printf("load average: %v %v %v\n", stats.Load1, stats.Load5, stats.Load10)
	fmt.Println("--------------------------")
}

func fmtBytes(val uint64) string {
	if val < 1024 {
		return fmt.Sprintf("%d bytes", val)
	} else if val < 1024*1024 {
		return fmt.Sprintf("%6.2f KiB", float64(val)/1024.0)
	} else if val < 1024*1024*1024 {
		return fmt.Sprintf("%6.2f MiB", float64(val)/1024.0/1024.0)
	} else {
		return fmt.Sprintf("%6.2f GiB", float64(val)/1024.0/1024.0/1024.0)
	}
}
