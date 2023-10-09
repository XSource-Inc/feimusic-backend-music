package db

// 这里写什么呢
import (
	"fmt"

	"github.com/Kidsunbo/kie_toolbox_go/logs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
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

func MustInit() error{
	var err error
	dsn := fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=utf8&parseTime=True",
		mysqlConfig.User,
		mysqlConfig.Password,
		mysqlConfig.Host,
		mysqlConfig.Port,
		mysqlConfig.Database,
)	

	db, err = gorm.Open("mysql", dsn)
	if err != nil{
		logs.Fatal("failed to connect database, err=%v", err)
		return err

	}

	if db.Error != nil{
		//没理解这里有两个error
		logs.Fatal("failed to connect database, err=%v", db.Error)
		return db.Error
	}

	return nil 

}
