package controllers

import (
	"hello/models"

	"github.com/astaxie/beego"
)

type BannerController struct {
	BaseController
}

func (c *BannerController) Index() {
	backstage := beego.URLFor("BackstageController.Index")
	id, _ := c.GetInt("id")
	// No (or an invalid) id means "show the first banner", preserving the
	// previous landing behavior.
	if id <= 0 {
		id = 1
	}
	banner, ok := c.loadBanner(id, backstage)
	if !ok {
		return
	}
	c.Data["Id"] = id
	c.Data["banner"] = banner
	c.TplName = "backstage/banner.html"
}

func (c *BannerController) Update() {
	backstage := beego.URLFor("BackstageController.Index")
	oldBanner, ok := c.requireBanner(backstage)
	if !ok {
		return
	}
	newbanner := new(models.Banner)
	flash := beego.NewFlash()
	newbanner.Id = oldBanner.Id
	newbanner.Title = c.GetString("title")
	newbanner.Subtitle = c.GetString("subtitle")
	newbanner.Url = c.GetString("url")
	file, image, err := c.GetFile("images")
	if err == nil {
		defer file.Close()
		newbanner.Imgurl = "static/upload/" + image.Filename
		if oldBanner.Imgurl != newbanner.Imgurl {
			err1 := c.SaveToFile("images", "static/upload/"+image.Filename) // 保存位置在 static/upload, 没有文件夹要先创建
			if err1 != nil {
				flash.Error("更新Banner失败！原因：" + err1.Error())
				flash.Store(&c.Controller)
				c.Data["Id"] = oldBanner.Id
				c.Data["banner"] = oldBanner
				c.TplName = "backstage/banner.html"
				return
			}
		}

	} else {
		newbanner.Imgurl = oldBanner.Imgurl
	}
	if err := newbanner.Update(); err != nil {
		flash.Error("更新Banner失败！原因：" + err.Error())
		flash.Store(&c.Controller)
		c.Data["Id"] = oldBanner.Id
		c.Data["banner"] = oldBanner
		c.TplName = "backstage/banner.html"
		return
	}

	flash.Error("更新成功！")
	flash.Store(&c.Controller)
	c.Data["Id"] = newbanner.Id
	c.Data["banner"] = newbanner
	c.TplName = "backstage/banner.html"
	return
}
