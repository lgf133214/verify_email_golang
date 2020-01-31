package main

import (
	_ "./routers"
	"./utils"
	"github.com/astaxie/beego"
)

func main() {
	go utils.AutoVerify()
	beego.Run()
}
