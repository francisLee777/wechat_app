package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
)

// oneLine compacts whitespace (\r, \n, \t, multiple spaces) into single spaces
func oneLine(s string) string {
	if s == "" {
		return s
	}
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	// collapse multiple spaces
	s = strings.Join(strings.Fields(s), " ")
	return s
}

func AccessLog() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()

		reqBody := c.Request.Body()
		reqBodyLog := ""
		ctReq := string(c.Request.Header.ContentType())
		if ctReq != "" && !strings.Contains(ctReq, "multipart/form-data") {
			reqBodyLog = string(reqBody)
			if len(reqBodyLog) > 500 {
				reqBodyLog = reqBodyLog[:500] + "..."
			}
		}

		c.Next(ctx)

		end := time.Now()
		latency := end.Sub(start)

		respBody := c.Response.Body()
		respBodyLog := ""
		ctResp := c.Response.Header.Get("Content-Type")
		path := string(c.Request.URI().PathOriginal())
		if ctResp != "" && (strings.HasPrefix(ctResp, "image/") || strings.HasPrefix(ctResp, "application/octet-stream")) {
			idx := strings.LastIndex(path, "/")
			fname := ""
			if idx >= 0 && idx+1 < len(path) {
				fname = path[idx+1:]
			}
			respBodyLog = "[image:" + fname + "]"
		} else if strings.HasPrefix(path, "/api/media/") {
			idx := strings.LastIndex(path, "/")
			fname := ""
			if idx >= 0 && idx+1 < len(path) {
				fname = path[idx+1:]
			}
			respBodyLog = "[image:" + fname + "]"
		} else {
			respBodyLog = oneLine(string(respBody))
			if len(respBodyLog) > 2000 {
				respBodyLog = respBodyLog[:2000] + "..."
			}
		}

		fmt.Printf("[Hertz] %v | %3d | %13v | %15s | %-7s %s\n[Request] %s\n[Response] %s\n",
			end.Format("2006/01/02 - 15:04:05"),
			c.Response.StatusCode(),
			latency,
			c.ClientIP(),
			string(c.Request.Header.Method()),
			path,
			reqBodyLog,
			respBodyLog,
		)
	}
}
