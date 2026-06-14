package models

import (
	"time"

	"github.com/astaxie/beego/orm"
)

type Blog struct {
	Id           int
	Auth         string
	Title        string
	Keywords     string
	Catalogid    string
	Content      string
	Introduction string
	Lastupdate   time.Time
	Type         string
	Status       string
	Views        string
	Imgurl       string
	Subject      string
	Createtime   time.Time `orm:"auto_now_add;type(datetime)"`
}

func (a *Blog) TableName() string {
	return "blog"
}

func GetBlogById(id int) (*Blog, error) {
	a := new(Blog)
	err := orm.NewOrm().QueryTable(TableName("blog")).Filter("id", id).One(a)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func BlogGetList(page, pageSize int, filters ...interface{}) ([]*Blog, int64) {
	offset := (page - 1) * pageSize
	list := make([]*Blog, 0)
	query := orm.NewOrm().QueryTable(TableName("blog"))
	//	fmt.Println(len(filters))
	if len(filters) > 1 {
		l := len(filters)
		for k := 0; k < l; k += 2 {
			query = query.Filter(filters[k].(string), filters[k+1])
		}
	}
	total, _ := query.Count()
	query.OrderBy("-id").Limit(pageSize, offset).All(&list)

	return list, total
}

func BlogAdd(a *Blog) (int64, error) {
	return orm.NewOrm().Insert(a)
}

func (a *Blog) Update(fields ...string) error {
	if _, err := orm.NewOrm().Update(a, fields...); err != nil {
		return err
	}
	return nil
}

// StatusPublic is the only article status allowed to appear on the public-facing
// pages. Every other value (e.g. "private", "protected", legacy "1" or the empty
// string) is treated as not publicly visible. Centralising this here is what lets
// the home page, archive, single page and related list share one consistent rule.
const StatusPublic = "public"

// IsPublicStatus reports whether an article carrying the given status may be shown
// on the public site.
func IsPublicStatus(status string) bool {
	return status == StatusPublic
}

// IsPublic reports whether this article may be shown on the public site.
func (a *Blog) IsPublic() bool {
	return a != nil && IsPublicStatus(a.Status)
}

// publicQuery returns a query already restricted to publicly visible articles.
// The bool is false when no "default" database is registered, so callers can
// degrade to empty results instead of letting orm.NewOrm() panic (e.g. in tests
// or when the database is unreachable).
func publicQuery() (orm.QuerySeter, bool) {
	if _, err := orm.GetDB("default"); err != nil {
		return nil, false
	}
	return orm.NewOrm().QueryTable(TableName("blog")).Filter("status", StatusPublic), true
}

// BlogGetPublicList is the single entry point for "articles allowed on the front
// end". It returns a page of publicly visible articles plus the total number of
// publicly visible articles matching the same filters. Because the count and the
// rows are produced by the very same query, paging totals and the rendered
// columns can never disagree.
func BlogGetPublicList(page, pageSize int, filters ...interface{}) ([]*Blog, int64) {
	list := make([]*Blog, 0)
	query, ok := publicQuery()
	if !ok {
		return list, 0
	}
	if len(filters) > 1 {
		l := len(filters)
		for k := 0; k+1 < l; k += 2 {
			query = query.Filter(filters[k].(string), filters[k+1])
		}
	}
	total, _ := query.Count()
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize
	query.OrderBy("-id").Limit(pageSize, offset).All(&list)
	return list, total
}

// GetPublicBlogById returns an article only when it exists AND is publicly
// visible. Invalid ids, missing rows, non-public rows and a missing database all
// yield an error, so the caller can redirect rather than render an empty page.
func GetPublicBlogById(id int) (*Blog, error) {
	if id <= 0 {
		return nil, orm.ErrNoRows
	}
	query, ok := publicQuery()
	if !ok {
		return nil, orm.ErrNoRows
	}
	a := new(Blog)
	if err := query.Filter("id", id).One(a); err != nil {
		return nil, err
	}
	return a, nil
}

// BlogGetRelated returns publicly visible articles from the same catalog as the
// article being viewed, never including that article itself.
func BlogGetRelated(catalogid string, excludeId, limit int) []*Blog {
	if limit <= 0 {
		return []*Blog{}
	}
	rows := make([]*Blog, 0)
	if query, ok := publicQuery(); ok {
		// Fetch one extra row so dropping the current article still leaves
		// enough related posts to fill the requested limit.
		query.Filter("catalogid", catalogid).OrderBy("-id").Limit(limit+1).All(&rows)
	}
	return selectRelated(rows, excludeId, limit)
}

// selectRelated reduces a list of candidate articles to the related posts to
// display: it drops the article currently being viewed and any non-public rows,
// and caps the result at limit. It is side-effect free so the exclusion and
// capping rules can be unit tested without a database.
func selectRelated(candidates []*Blog, excludeId, limit int) []*Blog {
	related := make([]*Blog, 0, limit)
	if limit <= 0 {
		return related
	}
	for _, b := range candidates {
		if b == nil || b.Id == excludeId || !b.IsPublic() {
			continue
		}
		related = append(related, b)
		if len(related) >= limit {
			break
		}
	}
	return related
}
