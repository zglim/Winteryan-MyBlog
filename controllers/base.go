package controllers

import (
	"fmt"
	"hello/libs"
	"hello/models"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
)

type BaseController struct {
	beego.Controller
	controllerName string
	actionName     string
	userId         int
	userName       string
	loginName      string
	pageSize       int
	allowUrl       string
}

//前期准备
func (self *BaseController) Prepare() {
	self.pageSize = 20
	controllerName, actionName := self.GetControllerAndAction()
	self.controllerName = strings.ToLower(controllerName[0 : len(controllerName)-10])
	self.actionName = strings.ToLower(actionName)
	self.Data["siteName"] = beego.AppConfig.String("site.name")
	self.Data["curRoute"] = self.controllerName + "." + self.actionName
	self.Data["curController"] = self.controllerName
	self.Data["curAction"] = self.actionName
	// noAuth := "ajaxsave/ajaxdel/table/loginin/loginout/getnodes/start"
	// isNoAuth := strings.Contains(noAuth, self.actionName)
	// if isNoAuth == false {
	fmt.Println(controllerName, self.actionName)
	self.auth()
	// }

	self.Data["loginUserId"] = self.userId
	self.Data["loginUserName"] = self.userName
}

//登录权限验证
func (self *BaseController) auth() {

	arr := strings.Split(self.Ctx.GetCookie("auth"), "|")
	self.userId = 0
	if len(arr) == 2 {
		idstr, password := arr[0], arr[1]
		userId, _ := strconv.Atoi(idstr)
		if userId > 0 {
			user, err := models.GetUserById(userId)
			if err == nil && password == libs.Md5([]byte(self.getClientIp()+"|"+user.Password+user.Personalkey)) {
				self.userId = user.Id
				self.loginName = user.Username
				self.userName = user.Username
			}
		}
	}

	if self.userId == 0 && (self.actionName != "login") {
		self.redirect(beego.URLFor("LoginController.Login"))
	}
}

func (self *BaseController) getClientIp() string {
	s := strings.Split(self.Ctx.Request.RemoteAddr, ":")
	return s[0]
}

// 重定向
func (self *BaseController) redirect(url string) {
	self.Redirect(url, 302)
	self.StopRun()
}

// 是否POST提交
func (self *BaseController) isPost() bool {
	return self.Ctx.Request.Method == "POST"
}

// 是否GET提交
func (self *BaseController) isGet() bool {
	return self.Ctx.Request.Method == "GET"
}

// flashAndRedirect 存储 flash 消息并重定向，然后终止当前请求
func (self *BaseController) flashAndRedirect(msg, url string) {
	flash := beego.NewFlash()
	flash.Error("%s", msg)
	flash.Store(&self.Controller)
	self.redirect(url)
}

// validateAndLoadBlog 校验 id 合法性并加载 Blog 记录；失败时自动 flash + redirect，返回 nil, false
func (self *BaseController) validateAndLoadBlog(id int, redirectURL string) (*models.Blog, bool) {
	if id <= 0 {
		self.flashAndRedirect("文章ID无效", redirectURL)
		return nil, false
	}
	blog, err := models.GetBlogById(id)
	if err != nil || blog == nil {
		self.flashAndRedirect("文章不存在或已被删除", redirectURL)
		return nil, false
	}
	return blog, true
}

// validateAndLoadBanner 校验 id 合法性并加载 Banner 记录；失败时自动 flash + redirect，返回 nil, false
func (self *BaseController) validateAndLoadBanner(id int, redirectURL string) (*models.Banner, bool) {
	if id <= 0 {
		self.flashAndRedirect("Banner ID无效", redirectURL)
		return nil, false
	}
	banner, err := models.GetBannerById(id)
	if err != nil || banner == nil {
		self.flashAndRedirect("Banner不存在或已被删除", redirectURL)
		return nil, false
	}
	return banner, true
}
