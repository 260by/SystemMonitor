package data

import (
    _ "github.com/mattn/go-sqlite3"
	"github.com/go-xorm/xorm"
	"github.com/go-xorm/core"
	"fmt"
	"strings"
	"strconv"
	"time"
	"golang.org/x/crypto/ssh"
	"github.com/260by/SystemMonitor/sys"
)

// User 用户数据表
type User struct {
	ID       int    `xorm:"pk autoincr notnull"`
	UserName     string `xorm:"varchar(32) notnull unique index"`
	Password string `xorm:"varchar(128) notnull"`
}

// Assets 资产数据表
type Assets struct {
	ID         int    `xorm:"pk autoincr notnull"`
	IP         string `xorm:"notnull"`
	Port	   int
	UserName   string `xorm:"varchar(32)"`
	Authenticate string `xorm:"text"`
}

// Monitor 监控数据表
type Monitor struct {
	ID             int   `xorm:"pk autoincr notnull"`
	CreateTime     int64 `xorm:"index"`
	HostName	   string `xorm:"varchar(64) index"`
	CPUUse         float64
	MemUse         float64
	DiskUse        string `xorm:"varchar(255)"`
	Load1          float64
	Load5          float64
	Load10         float64
	// TCPTimeWait    int
	// TCPEstablished int
	IP       string `xorm:"varchar(32) index"`
}

// Connect 连接数据库
func Connect(driveName, dataSourceName string, showSQL bool) (*xorm.Engine, error) {
	orm, err := xorm.NewEngine(driveName, dataSourceName)
	if err != nil {
		return nil, err
	}
	orm.SetMapper(core.GonicMapper{})
	orm.ShowSQL(showSQL)
	return orm, nil
}

// Migrate 同步表结构
func Migrate(orm *xorm.Engine,) error {
	err := orm.Sync2(&User{}, &Assets{}, &Monitor{})
	if err != nil {
		return err
	}
	return nil
}

// GetMonitorData 获取监控数据
func GetMonitorData(client *ssh.Client) Monitor {
	montior := Monitor{}

	stats := sys.Stats{}
	sys.GetAllStats(client, &stats)

	// var preCPU sys.CPURaw
	sys.GetCPU(client, &stats)

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
	montior.CPUUse = stats.CPUUse
	montior.MemUse = memUsePercent
	montior.DiskUse = diskUse
	montior.Load1, _ = strconv.ParseFloat(stats.Load1, 64)
	montior.Load5, _ = strconv.ParseFloat(stats.Load5, 64)
	montior.Load10, _ = strconv.ParseFloat(stats.Load10, 64)
	montior.IP = ip

	return montior
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