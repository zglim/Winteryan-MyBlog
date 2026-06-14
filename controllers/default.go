package controllers

import (
	"fmt"
	"hello/models"
	html "html/template"
	"math"
	"strconv"

	"github.com/astaxie/beego"
)

type MainController struct {
	beego.Controller
}

func (c *MainController) Get() {

	// Only show publicly visible articles on the homepage
	baseFilters := models.PublicBlogBaseFilters()
	homeFilters := models.MergeFilters(baseFilters, []interface{}{"type", "original"})
	r1, total := models.BlogGetList(1, 4, homeFilters...)
	r2, _ := models.BlogGetList(2, 4, homeFilters...)
	r3, _ := models.BlogGetList(3, 4, homeFilters...)
	fmt.Println(r2)
	fmt.Println(r3)
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
	if page == 0 {
		page = 1
	}
	// Archive shows only publicly visible articles
	baseFilters := models.PublicBlogBaseFilters()
	r1, total := models.BlogGetList(page*3-2, 4, baseFilters...)
	r2, _ := models.BlogGetList(page*3-1, 4, baseFilters...)
	r3, _ := models.BlogGetList(page*3, 4, baseFilters...)
	fmt.Println(r2)
	fmt.Println(r3)
	pages := (int)(math.Ceil(float64(total) / float64(12)))
	filters2 := make([]interface{}, 0)
	for a := 1; a <= pages; a++ {
		var tempStr = "<a href=\"/archive?page=" + strconv.Itoa(a) + "\">" + strconv.Itoa(a) + "</a>"
		filters2 = append(filters2, html.HTML(tempStr))
	}
	fmt.Println(filters2)
	c.Data["Pages"] = filters2
	c.Data["List1"] = r1
	c.Data["List2"] = r2
	c.Data["List3"] = r3
	c.Data["Total"] = total

	c.TplName = "portals/archive.html"
}

func (main *MainController) Single() {
	id, _ := main.GetInt("id")
	if id <= 0 {
		main.Redirect(beego.URLFor("MainController.Archive"), 302)
		main.StopRun()
		return
	}
	// Use GetPublicBlogById so private/non-existent articles are blocked
	blog, err := models.GetPublicBlogById(id)
	if err != nil || blog == nil {
		main.Redirect(beego.URLFor("MainController.Archive"), 302)
		main.StopRun()
		return
	}
	// Related articles: same catalog, public only, exclude current article
	baseFilters := models.PublicBlogBaseFilters()
	relatedFilters := models.MergeFilters(baseFilters, []interface{}{"catalogid", blog.Catalogid})
	r1, _ := models.BlogGetList(1, 3, relatedFilters...)
	// Exclude current article from related list
	related := make([]*models.Blog, 0, len(r1))
	for _, b := range r1 {
		if b.Id != blog.Id {
			related = append(related, b)
		}
	}
	main.Data["Related"] = related
	main.Data["Blog"] = blog
	main.TplName = "portals/single.html"
}

func (main *MainController) Contact() {
	main.TplName = "portals/contact.html"
}
