package controllers

import (
	"fmt"
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
	file, image, err := c.GetFile("images")
	if err == nil {
		defer file.Close()
		newbanner.Imgurl = "static/upload/" + image.Filename
		fmt.Println(newbanner.Imgurl, oldBanner.Imgurl)
		if oldBanner.Imgurl != newbanner.Imgurl {
			err1 := c.SaveToFile("images", "static/upload/"+image.Filename) // 保存位置在 static/upload, 没有文件夹要先创建
			if err1 != nil {
				flash.Error("%s", "更新Banner失败！原因：" + err1.Error())
				flash.Store(&c.Controller)
				c.Data["banner"] = oldBanner
				c.TplName = "backstage/banner.html"
				return
			}
		}

	} else {
		newbanner.Imgurl = oldBanner.Imgurl
	}
	if err := newbanner.Update(); err != nil {
		flash.Error("%s", "更新Banner失败！原因：" + err.Error())
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
