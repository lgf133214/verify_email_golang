package models

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	Id int64
	Email string
	VerifyTime string
}

func init() {
	driverName := "mysql"
	user := beego.AppConfig.String("user")
	passwd := beego.AppConfig.String("passwd")
	host := beego.AppConfig.String("host")
	port := beego.AppConfig.String("port")
	database := beego.AppConfig.String("database")
	dbconn := user + ":" + passwd + "@tcp(" + host + ":" + port + ")/" + database + "?charset=utf8"
	err := orm.RegisterDataBase("default", driverName, dbconn)
	if err != nil {
		logs.Error("数据库连接错误")
		return
	}
	logs.Info("数据库连接成功")
	orm.RegisterModel(new(User))
	orm.RunSyncdb("default", false, true)
}