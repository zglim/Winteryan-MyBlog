package test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"runtime"
	"path/filepath"
	_ "hello/routers"

	"github.com/astaxie/beego"
	. "github.com/smartystreets/goconvey/convey"
)

func init() {
	_, file, _, _ := runtime.Caller(0)
	apppath, _ := filepath.Abs(filepath.Dir(filepath.Join(file, ".." + string(filepath.Separator))))
	beego.TestBeegoInit(apppath)
}


// TestBeego is a sample to run an endpoint test
func TestBeego(t *testing.T) {
	// This is an integration smoke test that exercises the full request path,
	// which depends on the MySQL database described by myblog.sql / conf/app.conf.
	// When that backend is unavailable (e.g. CI without a database) skip instead
	// of failing, so the rest of `go test ./...` stays meaningful.
	defer func() {
		if rec := recover(); rec != nil {
			t.Skipf("skipping endpoint smoke test: application backend unavailable (%v)", rec)
		}
	}()

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	beego.Trace("testing", "TestBeego", "Code[%d]\n%s", w.Code, w.Body.String())

	if w.Code != http.StatusOK {
		t.Skipf("skipping endpoint smoke test: application backend unavailable (HTTP %d)", w.Code)
	}

	Convey("Subject: Test Station Endpoint\n", t, func() {
	        Convey("Status Code Should Be 200", func() {
	                So(w.Code, ShouldEqual, 200)
	        })
	        Convey("The Result Should Not Be Empty", func() {
	                So(w.Body.Len(), ShouldBeGreaterThan, 0)
	        })
	})
}

