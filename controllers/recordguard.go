package controllers

import (
	"hello/models"

	"github.com/astaxie/beego"
)

// recordguard.go centralizes how the backstage reacts to requests that
// reference a record by id.
//
// Several article/banner actions used to do `obj, _ := models.GetXById(id)`
// and then immediately dereference the result (obj.Imgurl, obj.Status, ...).
// Whenever the id was missing, malformed, or pointed at a row that no longer
// existed, GetXById returned (nil, err); the ignored error meant the action
// either panicked on the nil pointer or silently carried a zero-value object
// into the template. The helpers below give all of those cases one
// predictable behavior: store a human-readable flash message and redirect
// back to a safe page, returning ok=false so the caller can simply `return`
// instead of continuing with a nil/zero object.

// notFound stores msg as a flash error and redirects to fallback. It is the
// shared building block for every "record does not exist" branch so the
// behavior stays consistent across actions. Because redirect() calls
// StopRun(), execution does not return to the caller in the normal beego
// request flow.
func (self *BaseController) notFound(msg, fallback string) {
	flash := beego.NewFlash()
	flash.Error(msg)
	flash.Store(&self.Controller)
	self.redirect(fallback)
}

// requireID reads and validates the "id" parameter of the current request
// (works for both GET query and POST form values). A missing, non-numeric or
// non-positive id is treated as "no such record": it triggers
// notFound(fallback) and returns ok=false.
func (self *BaseController) requireID(fallback string) (int, bool) {
	id, err := self.GetInt("id")
	if err != nil || id <= 0 {
		self.notFound("操作失败！记录ID缺失或不合法。", fallback)
		return 0, false
	}
	return id, true
}

// loadBlog loads the blog with the given id. A missing row (or any query
// error) is funneled through notFound(fallback) and returns (nil, false).
func (self *BaseController) loadBlog(id int, fallback string) (*models.Blog, bool) {
	blog, err := models.GetBlogById(id)
	if err != nil || blog == nil {
		self.notFound("操作失败！未找到对应的文章，它可能已被删除。", fallback)
		return nil, false
	}
	return blog, true
}

// requireBlog resolves the blog referenced by the request id. It returns
// (blog, true) only when the id is valid and the row exists; otherwise a
// flash has already been stored and a redirect to fallback issued, and it
// returns (nil, false).
func (self *BaseController) requireBlog(fallback string) (*models.Blog, bool) {
	id, ok := self.requireID(fallback)
	if !ok {
		return nil, false
	}
	return self.loadBlog(id, fallback)
}

// loadBanner mirrors loadBlog for banner records. It is exposed separately so
// callers that resolve the id themselves (e.g. the banner index page, which
// falls back to banner #1 when no id is supplied) can still reuse the shared
// not-found handling.
func (self *BaseController) loadBanner(id int, fallback string) (*models.Banner, bool) {
	banner, err := models.GetBannerById(id)
	if err != nil || banner == nil {
		self.notFound("操作失败！未找到对应的Banner，它可能已被删除。", fallback)
		return nil, false
	}
	return banner, true
}

// requireBanner mirrors requireBlog for banner records.
func (self *BaseController) requireBanner(fallback string) (*models.Banner, bool) {
	id, ok := self.requireID(fallback)
	if !ok {
		return nil, false
	}
	return self.loadBanner(id, fallback)
}
