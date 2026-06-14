package controllers

import (
	"hello/libs"
	"hello/models"

	"github.com/astaxie/beego"
)

type BannerController struct {
	BaseController
}

func (c *BannerController) Index() {
	id, _ := c.GetInt("id")
	if id == 0 {
		id = 1
	}
	banner, _ := models.GetBannerById(id)
	c.Data["Id"] = id
	c.Data["banner"] = banner
	c.TplName = "backstage/banner.html"
}

func (c *BannerController) Update() {
	newbanner := new(models.Banner)
	id, _ := c.GetInt("id")
	flash := beego.NewFlash()
	oldBanner, _ := models.GetBannerById(id)
	newbanner.Id = id
	newbanner.Title = c.GetString("title")
	newbanner.Subtitle = c.GetString("subtitle")
	newbanner.Url = c.GetString("url")
	savedPath, changed, err := c.saveUploadImage("images")
	if err != nil {
		flash.Error("更新Banner失败！原因：" + err.Error())
		flash.Store(&c.Controller)
		c.Data["banner"] = oldBanner
		c.TplName = "backstage/banner.html"
		return
	}
	newbanner.Imgurl = libs.ResolveImageURL(oldBanner.Imgurl, savedPath, changed)
	if err := newbanner.Update(); err != nil {
		flash.Error("更新Banner失败！原因：" + err.Error())
		flash.Store(&c.Controller)
		c.Data["banner"] = oldBanner
		c.TplName = "backstage/banner.html"
		return
	}

	flash.Error("更新成功！")
	flash.Store(&c.Controller)
	c.Data["banner"] = newbanner
	c.TplName = "backstage/banner.html"
	return
}
