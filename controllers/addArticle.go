package controllers

import (
	"fmt"
	"hello/libs"
	"hello/models"
	html "html/template"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"
)

type AddarticleController struct {
	BaseController
}

func (c *AddarticleController) Add() {
	beego.ReadFromRequest(&c.Controller)
	if c.isPost() {
		Blog := new(models.Blog)
		flash := beego.NewFlash()
		Blog.Title = strings.TrimSpace(c.GetString("title"))
		Blog.Keywords = strings.TrimSpace(c.GetString("keyword"))
		Blog.Subject = strings.TrimSpace(c.GetString("subject"))
		Blog.Catalogid = strings.TrimSpace(c.GetString("catalogid"))
		Blog.Type = strings.TrimSpace(c.GetString("type"))
		Blog.Status = strings.TrimSpace(c.GetString("status"))
		Blog.Auth = strings.TrimSpace(c.GetString("auth"))
		Blog.Introduction = strings.TrimSpace(c.GetString("introduction"))
		Blog.Content = strings.TrimSpace(c.GetString("content"))
		Blog.Lastupdate = time.Now()
		Blog.Createtime = time.Now()
		savedPath, changed, err := c.saveUploadImage("images")
		if err != nil {
			flash.Error("保存Blog失败！原因：" + err.Error())
			flash.Store(&c.Controller)
			c.redirect(beego.URLFor("AddarticleController.Add"))
			return
		}
		if !changed {
			flash.Error("保存Blog失败！原因：请上传封面图片")
			flash.Store(&c.Controller)
			c.redirect(beego.URLFor("AddarticleController.Add"))
			return
		}
		Blog.Imgurl = savedPath
		if _, err := models.BlogAdd(Blog); err != nil {
			flash.Error("保存Blog失败！原因：" + err.Error())
			flash.Store(&c.Controller)
			c.redirect(beego.URLFor("AddarticleController.Add"))
			return
		}
		flash.Error("保存成功！")
		flash.Store(&c.Controller)
		c.redirect(beego.URLFor("AddarticleController.Add"))
		return

	}
	c.TplName = "backstage/addarticle.html"
}
func (c *AddarticleController) List() {
	page, _ := strconv.Atoi(strings.TrimSpace(c.GetString("page")))
	//	fmt.Println(page + "+++++++++++++++++++++++++++")
	filters := make([]interface{}, 0)
	r1, total := models.BlogGetList(page, 10, filters...)
	c.Data["List"] = r1
	pages := (int)(math.Ceil(float64(total) / float64(10)))
	fmt.Println(pages)
	filters2 := make([]interface{}, 0)
	for a := 1; a <= pages; a++ {
		var tempStr = "<a href=\"/listArticle?page=" + strconv.Itoa(a) + "\">" + strconv.Itoa(a) + "</a>"
		filters2 = append(filters2, html.HTML(tempStr))
	}
	fmt.Println(filters2)
	c.Data["Pages"] = filters2
	c.TplName = "backstage/listarticle.html"
}

func (c *AddarticleController) Update() {
	if c.isGet() {
		id, _ := c.GetInt("id")
		fmt.Println(id)
		blog, _ := models.GetBlogById(id)
		c.Data["blog"] = blog
		c.TplName = "backstage/addarticle.html"
	}
	if c.isPost() {
		Blog := new(models.Blog)
		id, _ := c.GetInt("id")
		flash := beego.NewFlash()
		oldblog, _ := models.GetBlogById(id)
		Blog.Id = id
		Blog.Auth = c.GetString("auth")
		Blog.Catalogid = c.GetString("catalogid")
		Blog.Content = c.GetString("content")
		Blog.Introduction = c.GetString("introduction")
		Blog.Keywords = c.GetString("keyword")
		Blog.Lastupdate = time.Now()
		Blog.Status = c.GetString("status")
		Blog.Subject = c.GetString("subject")
		Blog.Title = c.GetString("title")
		Blog.Type = c.GetString("type")
		savedPath, changed, err := c.saveUploadImage("images")
		if err != nil {
			flash.Error("更新Blog失败！原因：" + err.Error())
			flash.Store(&c.Controller)
			c.Data["blog"] = oldblog
			c.TplName = "backstage/addarticle.html"
			return
		}
		Blog.Imgurl = libs.ResolveImageURL(oldblog.Imgurl, savedPath, changed)
		if err := Blog.Update(); err != nil {
			flash.Error("更新Blog失败！原因：" + err.Error())
			flash.Store(&c.Controller)
			c.Data["blog"] = oldblog
			c.TplName = "backstage/addarticle.html"
			return
		}

		flash.Error("更新成功！")
		flash.Store(&c.Controller)
		newblog, _ := models.GetBlogById(id)
		c.Data["blog"] = newblog
		c.Data["flag"] = "1"
		c.TplName = "backstage/addarticle.html"
		return
	}

}
func (c *AddarticleController) Look() {
	id, _ := c.GetInt("id")
	fmt.Println(id)
	blog, _ := models.GetBlogById(id)
	c.Data["blog"] = blog
	c.Data["flag"] = "1"
	c.TplName = "backstage/addarticle.html"

}
func (c *AddarticleController) Delete() {
	id, _ := c.GetInt("id")
	fmt.Println(id)
	flash := beego.NewFlash()
	blog, _ := models.GetBlogById(id)
	blog.Status = "private"
	err := blog.Update()
	if err != nil {
		flash.Error("修改失败！原因：" + err.Error())
		flash.Store(&c.Controller)
		c.redirect(beego.URLFor("AddarticleController.List"))
	}
	flash.Error("修改成功！")
	flash.Store(&c.Controller)
	c.redirect(beego.URLFor("AddarticleController.List"))
}
