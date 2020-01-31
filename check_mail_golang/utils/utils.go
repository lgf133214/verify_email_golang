package utils

import (
	"../models"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/pkg/errors"
	"net"
	"net/smtp"
	"regexp"
	"strings"
	"time"
)

const forceDisconnectAfter = time.Second * 5

var emailRegexp = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func init() {
	logs.SetLogger(logs.AdapterFile, `{"filename":"./log.log"}`)
}

func RegexMatch(email string) error {
	if !emailRegexp.MatchString(email) {
		return errors.New("邮箱格式错误，请检查后再试")
	}
	return nil
}

func VerifyHost(email string) error {
	_, host := split(email)
	mx, err := net.LookupMX(host)
	if err != nil {
		return errors.New("邮箱域名错误，请检查后再试")
	}

	client, err := DialTimeout(fmt.Sprintf("%s:%d", mx[0].Host, 25), forceDisconnectAfter)
	if err != nil {
		return errors.New("邮箱验证失败，请检查后再试")
	}
	defer client.Close()

	err = client.Hello("checkmail.me")
	if err != nil {
		return errors.New("邮箱验证失败，请检查后再试")
	}
	err = client.Mail("lansome-cowboy@gmail.com")
	if err != nil {
		return errors.New("邮箱验证失败，请检查后再试")
	}
	err = client.Rcpt(email)
	if err != nil {
		return errors.New("邮箱验证失败，请检查后再试")
	}
	return nil
}

func DialTimeout(addr string, timeout time.Duration) (*smtp.Client, error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}

	t := time.AfterFunc(timeout, func() { conn.Close() })
	defer t.Stop()

	host, _, _ := net.SplitHostPort(addr)
	return smtp.NewClient(conn, host)
}

func split(email string) (account, host string) {
	i := strings.LastIndexByte(email, '@')
	account = email[:i]
	host = email[i+1:]
	return
}

// 定时检验
func AutoVerify() {
	timeUpdate, _:=beego.AppConfig.Int("autoVerifyTime")
	o := orm.NewOrm()
	var tmp models.User
	for {
		locker := make(chan struct{}, 1)
		locker <- struct{}{}
		go func() {
			now := time.Now().Day()
			o.Raw("select * from user u order by rand() limit 1;").QueryRow(&tmp)
			day, _ := time.ParseInLocation("2006-01-02 15:04:05", tmp.VerifyTime, time.Local)
			if now != day.Day() {
				verifyAll()
			}
			<-locker
		}()
		time.Sleep(time.Hour*time.Duration(timeUpdate))
	}

}

func verifyAll() {
	o := orm.NewOrm()
	var rows []models.User
	o.Raw("select * from user;").QueryRows(&rows)
	for _, val := range rows {
		email := val.Email
		err := VerifyHost(email)
		if err != nil {
			// 删除这个字段
			_, err := o.Delete(&val)
			if err != nil {
				logs.Error("删除失败")
			}
		} else {
			// 更新时间
			val.VerifyTime = time.Now().Format("2006-01-02 15:04:05")
			_, err := o.Update(&val)
			if err != nil {
				logs.Error("更新失败")
			}
		}
	}
	logs.Info("更新成功")
}
