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

const articlePageSize = 10

// articleRow wraps a Blog together with the fully-resolved operation URLs so
// the template can render real links without relying on client-side scripts.
type articleRow struct {
	*models.Blog
	LookURL   string
	EditURL   string
	DeleteURL string
}

// pageLink is a single numbered entry in the pagination bar.
type pageLink struct {
	Num    int
	URL    string
	Active bool
}

// pagination holds everything the template needs to render the pager,
// including the previous/next controls and their enabled state.
type pagination struct {
	Items      []pageLink
	HasPrev    bool
	PrevURL    string
	HasNext    bool
	NextURL    string
	Current    int
	TotalPages int
}

// articleActionURL builds the server-side URL for a per-row operation so the
// view, edit and delete buttons work even when JavaScript is disabled.
func articleActionURL(action string, id int) string {
	var path string
	switch action {
	case "look":
		path = "/lookArticle"
	case "update":
		path = "/updateArticle"
	case "delete":
		path = "/deleteArticle"
	default:
		path = "/listArticle"
	}
	return path + "?id=" + strconv.Itoa(id)
}

// articlePageURL builds the URL for a given page of the list.
func articlePageURL(page int) string {
	return "/listArticle?page=" + strconv.Itoa(page)
}

// normalizePage clamps a requested page into the valid [1, totalPages] range so
// missing, zero, negative or out-of-range page values all resolve to a real
// page instead of producing a negative offset or an empty first page.
func normalizePage(page, totalPages int) int {
	if totalPages < 1 {
		totalPages = 1
	}
	if page < 1 {
		return 1
	}
	if page > totalPages {
		return totalPages
	}
	return page
}

// buildPagination assembles the pager for the given current page, wiring up the
// previous/next links and the active highlight. The current page is clamped so
// the first/last boundaries are always handled correctly.
func buildPagination(current, totalPages int) pagination {
	if totalPages < 1 {
		totalPages = 1
	}
	current = normalizePage(current, totalPages)

	p := pagination{
		Current:    current,
		TotalPages: totalPages,
		HasPrev:    current > 1,
		HasNext:    current < totalPages,
	}
	if p.HasPrev {
		p.PrevURL = articlePageURL(current - 1)
	}
	if p.HasNext {
		p.NextURL = articlePageURL(current + 1)
	}
	items := make([]pageLink, 0, totalPages)
	for i := 1; i <= totalPages; i++ {
		items = append(items, pageLink{
			Num:    i,
			URL:    articlePageURL(i),
			Active: i == current,
		})
	}
	p.Items = items
	return p
}

func (c *AddarticleController) List() {
	page, _ := strconv.Atoi(strings.TrimSpace(c.GetString("page")))
	if page < 1 {
		page = 1
	}
	filters := make([]interface{}, 0)
	list, total := models.BlogGetList(page, articlePageSize, filters...)
	totalPages := int(math.Ceil(float64(total) / float64(articlePageSize)))
	if totalPages < 1 {
		totalPages = 1
	}
	if page > totalPages {
		page = totalPages
		list, _ = models.BlogGetList(page, articlePageSize, filters...)
	}

	rows := make([]articleRow, 0, len(list))
	for _, b := range list {
		rows = append(rows, articleRow{
			Blog:      b,
			LookURL:   articleActionURL("look", b.Id),
			EditURL:   articleActionURL("update", b.Id),
			DeleteURL: articleActionURL("delete", b.Id),
		})
	}
	c.Data["List"] = rows
	c.Data["Pagination"] = buildPagination(page, totalPages)
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
