package data

import (
	"fmt"
	"strings"
	"strconv"
	"time"
	"golang.org/x/crypto/ssh"
	"github.com/jinzhu/gorm"
	"github.com/260by/SystemMonitor/sys"
)

const cpuDefaultRefresh = 3  // 以秒为单位的默认刷新CPU间隔

// InsertData 插入数据到数据库
func InsertData(client *ssh.Client, db *gorm.DB) {
	montior := Monitor{}

	stats := sys.Stats{}
	sys.GetAllStats(client, &stats)

	var preCPU sys.CPURaw
	sys.GetCPU(client, &stats, preCPU, cpuDefaultRefresh)

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