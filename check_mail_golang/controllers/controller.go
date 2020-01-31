package controllers

import (
	"../models"
	"../utils"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"time"
)

type ResponseData struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type MainController struct {
	beego.Controller
}

func (m *MainController) Get() {
	email := m.Ctx.Input.Param(":email")
	// 第一步：邮箱规则验证
	err := utils.RegexMatch(email)
	if err != nil {
		res := ResponseData{Code: 0, Message: err.Error()}
		m.Data["json"] = res
		m.ServeJSON()
		return
	}
	// 第二步：真实性验证

	// 先去数据库匹配
	o := orm.NewOrm()

	// 这种方法不可行，因为id的默认值是0，坑啊
	//user := models.User{Email: email}
	//err = o.Read(&user)

	var user models.User
	o.Raw("select * from user where email=?;", email).QueryRow(&user)
	if user.Email!= "" {
		// 数据库中存在则直接返回正确
		res := ResponseData{Code: 1, Message: "验证成功"}
		m.Data["json"] = res
		m.ServeJSON()
		return
	}

	err = utils.VerifyHost(email)
	if err != nil {
		res := ResponseData{Code: 0, Message: err.Error()}
		m.Data["json"] = res
		m.ServeJSON()
	} else {
		// 验证通过，返回结果
		res := ResponseData{Code: 1, Message: "验证成功"}
		m.Data["json"] = res
		m.ServeJSON()
		// 插入数据库
		_, err = o.Insert(&models.User{Email: email, VerifyTime: time.Now().Format("2006-01-02 15:04:05")})
		if err != nil {
			logs.Error("插入失败")
		}
	}
}
