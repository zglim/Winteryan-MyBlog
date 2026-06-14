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
	if getBlogByIdFunc != nil {
		return getBlogByIdFunc(id)
	}
	a := new(Blog)
	err := orm.NewOrm().QueryTable(TableName("blog")).Filter("id", id).One(a)
	if err != nil {
		return nil, err
	}
	return a, nil
}

// getBlogByIdFunc 仅供测试注入使用，生产环境为 nil
var getBlogByIdFunc func(int) (*Blog, error)

// SetGetBlogByIdFunc 用于测试时替换 GetBlogById 行为，传 nil 恢复默认
func SetGetBlogByIdFunc(f func(int) (*Blog, error)) {
	getBlogByIdFunc = f
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
