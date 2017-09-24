package routers

import (
	"weilin/controllers"
	"github.com/astaxie/beego"
)

func init() {
	yxcheckinthrow := beego.NewNamespace("/v1",
		beego.NSAutoRouter(&controllers.MainController{}),
	)
	beego.AddNamespace(yxcheckinthrow)
}
