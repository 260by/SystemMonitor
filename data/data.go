package data

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// User 用户数据表
type User struct {
	ID       int    `gorm:"primary_key;AUTO_INCREMENT"`
	Name     string `gorm:"type:varchar(32)"`
	Password string `gorm:"type:varchar(128)"`
}

// Assets 资产数据表
type Assets struct {
	ID         int    `gorm:"primary_key;AUTO_INCREMENT"`
	IP         string `gorm:"type:varchar(32)"`
	Port	   int
	UserName   string `gorm:"type:varchar(32)"`
	PrivateKey string `gorm:"type:text"`
}

// Monitor 监控数据表
type Monitor struct {
	ID             int   `gorm:"primary_key;AUTO_INCREMENT"`
	CreateTime     int64 `gorm:"index"`
	HostName	   string `gorm:"type:varchar(64);index"`
	CPUUse         float32
	MemUse         float64
	DiskUse        string `gorm:"type:varchar(32)"`
	Load1          float64
	Load5          float64
	Load10         float64
	// TCPTimeWait    int
	// TCPEstablished int
	IP       string `gorm:"type:varchar(32);index"`
}

// Conn 连接数据库
func Conn() (*gorm.DB, error) {
	db, err := gorm.Open("sqlite3", "monitor.db")
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&User{}, &Assets{}, &Monitor{})

	return db, nil
}
