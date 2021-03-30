package common

import (
	"net/http"
	"strings"
)

// 声明一个新的数据类型(函数类型)
type FilterHandle func(rw http.ResponseWriter, req *http.Request) error

// 拦截器结构体
type Filter struct {
	// 用来存储要拦截的uri
	filterMap map[string]FilterHandle
}

// Filter初始化函数
func NewFilter() *Filter {
	return &Filter{filterMap: make(map[string]FilterHandle)}
}

// 注册拦截器
func (f *Filter) RegisterFilterUri(uri string, handle FilterHandle) {
	f.filterMap[uri] = handle
}

// 根据uri获取对应的handler
func (f *Filter) GetFilterHandle(uri string) FilterHandle {
	return f.filterMap[uri]
}

// 声明新的函数类型
type WebHandle func(rw http.ResponseWriter, req *http.Request)

// 执行拦截器，返回函数类型
func (f *Filter) Handle(webHandle WebHandle) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		for path, handle := range f.filterMap {
			if strings.Contains(req.RequestURI, path) {
				//执行拦截业务逻辑
				err := handle(rw, req)
				if err != nil {
					rw.Write([]byte(err.Error()))
					return
				}
				// 跳出循环
				break
			}
		}
		// 执行正常注册的函数
		webHandle(rw, req)
	}

}
