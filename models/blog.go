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

// PublicBlogBaseFilters returns the base filter conditions for publicly visible blogs.
// Only articles with status="public" should be visible on the front-end.
func PublicBlogBaseFilters() []interface{} {
	return []interface{}{"status", "public"}
}

// MergeFilters merges multiple filter slices into one.
func MergeFilters(filterSets ...[]interface{}) []interface{} {
	merged := make([]interface{}, 0)
	for _, fs := range filterSets {
		merged = append(merged, fs...)
	}
	return merged
}

// GetPublicBlogById returns a blog only if it exists and is publicly visible.
func GetPublicBlogById(id int) (*Blog, error) {
	a := new(Blog)
	err := orm.NewOrm().QueryTable(TableName("blog")).
		Filter("id", id).
		Filter("status", "public").
		One(a)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (a *Blog) Update(fields ...string) error {
	if _, err := orm.NewOrm().Update(a, fields...); err != nil {
		return err
	}
	return nil
}
