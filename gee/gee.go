package gee

import (
	"log"
	"net/http"
	"path"
	"regexp"
	"strings"
)

// 方法定义，给url绑定func
type HandlerFunc func(c *Context)

type Engine struct {
	*RouterGroup //作为顶层分组
	router       *router
	groups       []*RouterGroup //存储所有的group
}

// 分组结构
type RouterGroup struct {
	prefix      string        //分组url前缀
	middlewares []HandlerFunc //分组中间件
	parent      *RouterGroup  //父级分组
	engine      *Engine       //全局engine
}

// 新建一个实例
func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

// 添加group
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

// 添加新路由
func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	reg, _ := regexp.Compile("/+")
	pattern = string(reg.ReplaceAll([]byte(pattern), []byte("/")))
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.AddRouter(method, pattern, handler)
}

// 添加GET
func (group *RouterGroup) GET(pattern string, hanlder HandlerFunc) {
	group.addRoute("GET", pattern, hanlder)
}

// 添加POST
func (group *RouterGroup) POST(pattern string, hanlder HandlerFunc) {
	group.addRoute("POST", pattern, hanlder)
}

// 启动服务器
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

// 绑定中间件
func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

// 转发规则
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	//匹配中间件
	var middlewares []HandlerFunc
	//url处理，去掉重复的/
	reg, _ := regexp.Compile("/+")
	req.URL.Path = string(reg.ReplaceAll([]byte(req.URL.Path), []byte("/")))

	for _, group := range engine.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(w, req)
	c.handlers = middlewares
	engine.router.handle(c)

}

// 创建静态资源handler,文件交给http库来处理
func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(group.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Wraiter, c.Req)
	}
}

// 静态资源处理方法注册
func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPath := path.Join(relativePath, "/*filepath")
	group.GET(urlPath, handler)
}
