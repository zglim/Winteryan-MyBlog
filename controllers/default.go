package controllers

import (
	"hello/models"
	html "html/template"
	"math"
	"strconv"

	"github.com/astaxie/beego"
)

type MainController struct {
	beego.Controller
}

// articlesPerColumn is how many articles each of the three front-end columns holds.
const articlesPerColumn = 4

// articlesPerPage is the number of articles shown on one home/archive page.
const articlesPerPage = articlesPerColumn * 3

func (c *MainController) Get() {
	// Only original-type articles that are publicly visible reach the home page.
	r1, total := models.BlogGetPublicList(1, articlesPerColumn, "type", "original")
	r2, _ := models.BlogGetPublicList(2, articlesPerColumn, "type", "original")
	r3, _ := models.BlogGetPublicList(3, articlesPerColumn, "type", "original")
	banners, _ := models.BannerGetList(1, 4)
	c.Data["List1"] = r1
	c.Data["List2"] = r2
	c.Data["List3"] = r3
	c.Data["Banners"] = banners
	c.Data["Total"] = total
	c.TplName = "portals/home.html"
}

func (c *MainController) Archive() {
	page, _ := c.GetInt("page")
	if page < 1 {
		page = 1
	}
	// Every column is restricted to publicly visible articles, and the page total
	// is taken from the same query so the pagination never lists empty pages.
	r1, total := models.BlogGetPublicList(page*3-2, articlesPerColumn)
	r2, _ := models.BlogGetPublicList(page*3-1, articlesPerColumn)
	r3, _ := models.BlogGetPublicList(page*3, articlesPerColumn)
	pages := int(math.Ceil(float64(total) / float64(articlesPerPage)))
	pageLinks := make([]interface{}, 0, pages)
	for a := 1; a <= pages; a++ {
		link := "<a href=\"/archive?page=" + strconv.Itoa(a) + "\">" + strconv.Itoa(a) + "</a>"
		pageLinks = append(pageLinks, html.HTML(link))
	}
	c.Data["Pages"] = pageLinks
	c.Data["List1"] = r1
	c.Data["List2"] = r2
	c.Data["List3"] = r3
	c.Data["Total"] = total
	c.TplName = "portals/archive.html"
}

func (main *MainController) Single() {
	id, err := main.GetInt("id")
	if err != nil || id <= 0 {
		main.Redirect(beego.URLFor("MainController.Archive"), 302)
		return
	}
	blog, err := models.GetPublicBlogById(id)
	if err != nil {
		// Missing, non-public or otherwise unavailable article: send the visitor
		// back to the archive instead of rendering an empty page.
		main.Redirect(beego.URLFor("MainController.Archive"), 302)
		return
	}
	main.Data["Related"] = models.BlogGetRelated(blog.Catalogid, blog.Id, 3)
	main.Data["Blog"] = blog
	main.TplName = "portals/single.html"
}

func (main *MainController) Contact() {
	main.TplName = "portals/contact.html"
}
