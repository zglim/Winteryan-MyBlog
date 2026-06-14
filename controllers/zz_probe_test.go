package controllers

import (
	"fmt"
	"runtime"
	"testing"

	"hello/models"

	"github.com/astaxie/beego/orm"
)

func TestProbeDB(t *testing.T) {
	// register an unreachable mysql alias
	err := orm.RegisterDataBase("default", "mysql", "root:nopass@tcp(127.0.0.1:3306)/nodb?charset=utf8")
	fmt.Println("RegisterDataBase err:", err)

	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("NewOrm PANICKED:", r)
			} else {
				fmt.Println("NewOrm OK (no panic)")
			}
		}()
		_ = orm.NewOrm()
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("GetBlogById PANICKED:", r)
			}
		}()
		b, e := models.GetBlogById(999)
		fmt.Println("GetBlogById ->", b, e)
	}()
}

func TestProbeCaller(t *testing.T) {
	_, f0, _, _ := runtime.Caller(0)
	fmt.Println("Caller(0) =", f0)
}
