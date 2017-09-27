package controllers

import (
	"github.com/astaxie/beego"
	"weilin/mysqlUtility"
	"crypto/sha1"
	"fmt"
	"weilin/mysqlUtility/dal"
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

	entity := dal.NewEntity(c.GetString("tbname"), "QP")
	entity["PageSize"] = c.GetString("PageSize")
	entity["PageIndex"] = c.GetString("PageIndex")
	entity["FieldsSelect"] = c.GetString("cols")
	entity["Condition"] = fmt.Sprintf("where %s", c.GetString("where"))

	data, _ := mysqlUtility.DContext.QueryPager(entity)

	c.Data["json"] = data
	c.ServeJSON()
}
