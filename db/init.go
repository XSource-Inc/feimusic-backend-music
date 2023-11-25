package db

import (
	"github.com/Kidsunbo/kie_toolbox_go/logs"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type mysqlModel struct {
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	User        string `yaml:"user"`
	Password    string `yaml:"password"`
	Database    string `yaml:"database"`
	MaxIdLeConn int    `yaml:"maxidleconn"`
	MaxOpenConn int    `yaml:"maxopenconn"`
	Debug       bool   `yaml:"debug"`
	IsPlural    bool   `yaml:"isplural"`
	TablePrefix string `yaml:"tableprefix"`
}

var db *gorm.DB
var mysqlConfig = mysqlModel{
	Host:        "127.0.0.1",
	Port:        3308,
	User:        "root",
	Password:    "root123456",
	Database:    "data_center",
	MaxIdLeConn: 10,
	MaxOpenConn: 100,
	Debug:       false,
	IsPlural:    true,
}

func MustInit() error {
	var err error

	dsn := "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		logs.Fatal("failed to connect database, err=%v", err)
		return err

	}

	if db.Error != nil {
		//没理解这里有两个error
		logs.Fatal("failed to connect database, err=%v", db.Error)
		return db.Error
	}

	return nil
}

func GetDB() *gorm.DB {
	return db
}
