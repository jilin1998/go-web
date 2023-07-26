package gee

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
)

// 异常恢复
func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%s", err)
				log.Printf("%s\n\n", trace(message))
				c.Fail(http.StatusInternalServerError, "Internal Server Error")
			}
		}()
		c.Next()
	}
}

// 将异常栈转换成字符串
func trace(message string) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:]) //跳过前3个栈
	var str strings.Builder
	str.WriteString(message + "\n Traceback:")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		funcName := fn.Name()
		str.WriteString(fmt.Sprintf("\n\t%s#%s:%d", file, funcName, line))
	}
	return str.String()
}
