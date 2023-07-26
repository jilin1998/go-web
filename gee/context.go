package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// 存储数据
type H map[string]interface{}

type Context struct {
	//请求响应
	Wraiter http.ResponseWriter
	Req     *http.Request
	//请求info
	Path       string
	Method     string
	Params     map[string]string
	StatusCode int
	//中间件
	handlers []HandlerFunc
	index    int

	//engine pointer
	engine *Engine
}

// 实例化一个新的Context
func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Path:    req.URL.Path,
		Method:  req.Method,
		Req:     req,
		Wraiter: w,
		index:   -1,
	}
}

// 下一个中间件
func (c *Context) Next() {
	c.index++
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

// 获取参数
func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

// 从表单取数据
func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

// 从url获取参数
func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

// 设置响应码
func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Wraiter.WriteHeader(code)
}

// 添加header
func (c *Context) SetHeader(key string, value string) {
	c.Wraiter.Header().Add(key, value)
}

// 返回字符串
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Wraiter.Write([]byte(fmt.Sprintf(format, values...)))
}

// 写入json
func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Wraiter)

	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Wraiter, err.Error(), 500)
	}

}

// 写入Data
func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Wraiter.Write(data)
}

// 直接结束，跳过后续中间件
func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.JSON(code, H{"message": err})
}

// 返回html
func (c *Context) HTML(code int, name string, data interface{}) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Wraiter, name, data); err != nil {
		c.Fail(500, err.Error())
	}
}
