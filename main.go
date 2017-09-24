package main

import (
	_ "weilin/routers"
	"github.com/astaxie/beego"
	"weilin/mysqlUtility"
)

func main() {
	mysqlUtility.StartInit()
	beego.Run()
}

