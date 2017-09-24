package controllers

import (
	"github.com/astaxie/beego"
	"weilin/mysqlUtility"
	"crypto/sha1"
	"fmt"
)

type MainController struct {
	beego.Controller
}

func (c *MainController) Get() {
	signid := sha1.New()
	signid.Write([]byte(beego.AppConfig.String("sha1::signid")))
	signid1:=signid.Sum(nil)
	if fmt.Sprintf("%x", signid1)!=c.GetString("signid"){
		c.Data["json"] = "签名不正确"
		c.ServeJSON()
		return
	}

	signname := sha1.New()
	signname.Write([]byte(beego.AppConfig.String("sha1::signname")))
	signname1:=signname.Sum(nil)
	if fmt.Sprintf("%x", signname1)!=c.GetString("signname"){
		c.Data["json"] = "签名不正确"
		c.ServeJSON()
		return
	}

	sql:="select * from "+c.GetString("Table")+" WHERE "+c.GetString("Column")+" = '"+c.GetString("Value")+"'"
	beego.Info(sql)
	data,_:=mysqlUtility.DContext.QueryData(sql)
	c.Data["json"] = data
	c.ServeJSON()
}
