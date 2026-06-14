package controllers

import (
	"fmt"
	"hello/libs"
	"hello/models"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
)

type LoginController struct {
	BaseController
}

func (c *LoginController) Login() {
	if c.userId > 0 {
		c.redirect(beego.URLFor("AddarticleController.Add"))
	}
	beego.ReadFromRequest(&c.Controller)
	if c.isPost() {
		username := strings.TrimSpace(c.GetString("username"))
		password := strings.TrimSpace(c.GetString("password"))
		//	c.Data["username"] = username
		//	c.Data["password"] = password
		filters := make([]interface{}, 0)
		filters = append(filters, "username", username)
		filters = append(filters, "password", password)
		filters = append(filters, "status", 1)
		userlist, total := models.UserGetList(1, 1, filters...)
		flash := beego.NewFlash()
		if total > 0 {
			user := userlist[0]
			authkey := libs.Md5([]byte(c.getClientIp() + "|" + user.Password + user.Personalkey))
			c.Ctx.SetCookie("auth", strconv.Itoa(user.Id)+"|"+authkey, 3600)
			fmt.Println(beego.URLFor("AddarticleController.Add") + "+++++++++++++++++++++++++")
			c.redirect(beego.URLFor("BackstageController.Index"))
		} else {
			flash.Error("%s", "输入信息有误，或账户已失效")
			c.Data["error"] = "输入信息有误，或账户已失效"
			flash.Store(&c.Controller)
			c.TplName = "backstage/login.html"
		}
	}
	c.TplName = "backstage/login.html"

}
