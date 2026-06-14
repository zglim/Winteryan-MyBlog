package controllers

import (
	"fmt"
	"hello/models"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"
)

// PageItem represents a single pagination link.
type PageItem struct {
	Num    int
	URL    string
	Active string
}

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
			flash.Error("%s", "保存Blog失败！原因："+err.Error())
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
			flash.Error("%s", "保存Blog失败！原因："+err1.Error())
			flash.Store(&c.Controller)
			c.redirect(beego.URLFor("AddarticleController.Add"))
			return
		}
		if _, err := models.BlogAdd(Blog); err != nil {
			flash.Error("%s", "保存Blog失败！原因："+err.Error())
			flash.Store(&c.Controller)
			c.redirect(beego.URLFor("AddarticleController.Add"))
			return
		}
		flash.Error("%s", "保存成功！")
		flash.Store(&c.Controller)
		c.redirect(beego.URLFor("AddarticleController.Add"))
		return

	}
	c.TplName = "backstage/addarticle.html"
}
// NormalizePage clamps a page number into [1, totalPages].
// totalPages must be >= 1.
func NormalizePage(page, totalPages int) int {
	if totalPages < 1 {
		totalPages = 1
	}
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}
	return page
}

// BuildPagination creates the list of page items used by the pagination bar.
func BuildPagination(currentPage, totalPages int) []PageItem {
	if totalPages < 1 {
		totalPages = 1
	}
	items := make([]PageItem, 0, totalPages)
	for i := 1; i <= totalPages; i++ {
		active := ""
		if i == currentPage {
			active = "active"
		}
		items = append(items, PageItem{
			Num:    i,
			URL:    "/listArticle?page=" + strconv.Itoa(i),
			Active: active,
		})
	}
	return items
}

func (c *AddarticleController) List() {
	page, _ := strconv.Atoi(strings.TrimSpace(c.GetString("page")))
	filters := make([]interface{}, 0)
	// First pass with page=1 to get total count; we'll re-query with the real page.
	_, total := models.BlogGetList(1, 10, filters...)
	totalPages := int(math.Ceil(float64(total) / float64(10)))
	if totalPages < 1 {
		totalPages = 1
	}
	page = NormalizePage(page, totalPages)

	filters2 := make([]interface{}, 0)
	r1, _ := models.BlogGetList(page, 10, filters2...)
	c.Data["List"] = r1

	prevPage := page - 1
	if prevPage < 1 {
		prevPage = 1
	}
	nextPage := page + 1
	if nextPage > totalPages {
		nextPage = totalPages
	}

	c.Data["Pages"] = BuildPagination(page, totalPages)
	c.Data["CurrentPage"] = page
	c.Data["TotalPages"] = totalPages
	c.Data["FirstPage"] = 1
	c.Data["LastPage"] = totalPages
	c.Data["PrevPage"] = prevPage
	c.Data["NextPage"] = nextPage

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
		file, image, err := c.GetFile("images")
		if err == nil {
			defer file.Close()
			Blog.Imgurl = "static/upload/" + image.Filename
			fmt.Println(Blog.Imgurl, oldblog.Imgurl)
			if oldblog.Imgurl != Blog.Imgurl {
				err1 := c.SaveToFile("images", "static/upload/"+image.Filename) // 保存位置在 static/upload, 没有文件夹要先创建
				if err1 != nil {
					flash.Error("%s", "更新Blog失败！原因："+err1.Error())
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
			flash.Error("%s", "更新Blog失败！原因："+err.Error())
			flash.Store(&c.Controller)
			c.Data["blog"] = oldblog
			c.TplName = "backstage/addarticle.html"
			return
		}

		flash.Error("%s", "更新成功！")
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
		flash.Error("%s", "修改失败！原因："+err.Error())
		flash.Store(&c.Controller)
		c.redirect(beego.URLFor("AddarticleController.List"))
	}
	flash.Error("%s", "修改成功！")
	flash.Store(&c.Controller)
	c.redirect(beego.URLFor("AddarticleController.List"))
}
