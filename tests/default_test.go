package test

import (
	"testing"
	"runtime"
	"path/filepath"
	_ "weilin/routers"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"strings"
	"sync"
)

func init() {
	_, file, _, _ := runtime.Caller(1)
	apppath, _ := filepath.Abs(filepath.Dir(filepath.Join(file, ".." + string(filepath.Separator))))
	beego.TestBeegoInit(apppath)
}

var okNub int
func TestBeego(t *testing.T) {
	var wg sync.WaitGroup
	okNub=0
	for i:=0;i<4000;i++ {
		go func() {
			wg.Add(1)
			testGO(&okNub, &wg)
		}()
	}
	wg.Wait()
	beego.Error(okNub)
}
func testPHP(okNub *int,wg *sync.WaitGroup)  {
	defer wg.Done()
	result,_:=httplib.Get("http://172.104.81.5/api/db.php?signid=a00559727a317e5f31fd8b2eb502be49cf48618d&signname=a94a8fe5ccb19ba61c4c0873d391e987982fbbd3&tbname=TypeId0&cols=*&where=valid=10").String()
	if strings.Contains(result,"15eb1bca9f11c7071f"){
		*okNub++
	}
}
func testGO(okNub *int,wg *sync.WaitGroup)  {
	defer wg.Done()
	result,_:=httplib.Get("http://localhost:7777/v1/Main/Get?Table=user&Column=name&Value=平凡&signid=a00559727a317e5f31fd8b2eb502be49cf48618d&signname=a94a8fe5ccb19ba61c4c0873d391e987982fbbd3").String()
	if strings.Contains(result,"verify_user_id"){
		*okNub++
	}
}

