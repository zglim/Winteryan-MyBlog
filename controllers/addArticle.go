package controllers

import (
	"fmt"
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
		file, image, err := c.GetFile("images")
		if err != nil {
			flash.Error("保存Blog失败！原因：" + err.Error())
			flash.Store(&c.Controller)
			c.redirect(beego.URLFor("AddarticleController.Add"))
			return
		}
		defer file.Close()
		Blog.Imgurl = "static/upload/" + image.Filename
		Blog.Introduction = strings.TrimSpace(c.GetString("introduction"))
		Blog.Content = strings.TrimSpace(c.GetString("content"))
		Blog.Lastupdate = time.Now()
		Blog.Createtime = time.Now()
		err1 := c.SaveToFile("images", "static/upload/"+image.Filename) // 保存位置在 static/upload, 没有文件夹要先创建
		if err1 != nil {
			flash.Error("保存Blog失败！原因：" + err1.Error())
			flash.Store(&c.Controller)
			c.redirect(beego.URLFor("AddarticleController.Add"))
			return
		}
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
	list := beego.URLFor("AddarticleController.List")
	if c.isGet() {
		blog, ok := c.requireBlog(list)
		if !ok {
			return
		}
		c.Data["blog"] = blog
		c.TplName = "backstage/addarticle.html"
	}
	if c.isPost() {
		oldblog, ok := c.requireBlog(list)
		if !ok {
			return
		}
		Blog := new(models.Blog)
		flash := beego.NewFlash()
		Blog.Id = oldblog.Id
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
		file, image, err := c.GetFile("images")
		if err == nil {
			defer file.Close()
			Blog.Imgurl = "static/upload/" + image.Filename
			if oldblog.Imgurl != Blog.Imgurl {
				err1 := c.SaveToFile("images", "static/upload/"+image.Filename) // 保存位置在 static/upload, 没有文件夹要先创建
				if err1 != nil {
					flash.Error("更新Blog失败！原因：" + err1.Error())
					flash.Store(&c.Controller)
					c.Data["blog"] = oldblog
					c.TplName = "backstage/addarticle.html"
					return
				}
			}

		} else {
			Blog.Imgurl = oldblog.Imgurl
		}
		if err := Blog.Update(); err != nil {
			flash.Error("更新Blog失败！原因：" + err.Error())
			flash.Store(&c.Controller)
			c.Data["blog"] = oldblog
			c.TplName = "backstage/addarticle.html"
			return
		}

		flash.Error("更新成功！")
		flash.Store(&c.Controller)
		// The row was just persisted, so it must exist; re-read it to show the
		// stored values, but fall back to the in-memory copy if the re-read
		// fails so the template never receives a nil object.
		newblog, err2 := models.GetBlogById(Blog.Id)
		if err2 != nil || newblog == nil {
			newblog = Blog
		}
		c.Data["blog"] = newblog
		c.Data["flag"] = "1"
		c.TplName = "backstage/addarticle.html"
		return
	}

}
func (c *AddarticleController) Look() {
	blog, ok := c.requireBlog(beego.URLFor("AddarticleController.List"))
	if !ok {
		return
	}
	c.Data["blog"] = blog
	c.Data["flag"] = "1"
	c.TplName = "backstage/addarticle.html"

}
func (c *AddarticleController) Delete() {
	list := beego.URLFor("AddarticleController.List")
	blog, ok := c.requireBlog(list)
	if !ok {
		return
	}
	flash := beego.NewFlash()
	blog.Status = "private"
	if err := blog.Update(); err != nil {
		flash.Error("修改失败！原因：" + err.Error())
		flash.Store(&c.Controller)
		c.redirect(list)
		return
	}
	flash.Error("修改成功！")
	flash.Store(&c.Controller)
	c.redirect(list)
}
