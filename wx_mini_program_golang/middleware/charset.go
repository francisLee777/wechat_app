package middleware

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
)

// CharsetUTF8 ensures responses include charset=utf-8 for text and json content types
func CharsetUTF8() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		c.Next(ctx)

		ct := c.Response.Header.Get("Content-Type")
		if ct == "" {
			// Default to JSON utf-8 when unspecified
			c.Response.Header.Set("Content-Type", "application/json; charset=utf-8")
			return
		}
		if strings.HasPrefix(ct, "application/json") && !strings.Contains(ct, "charset=") {
			c.Response.Header.Set("Content-Type", ct+"; charset=utf-8")
			return
		}
		if strings.HasPrefix(ct, "text/") && !strings.Contains(ct, "charset=") {
			c.Response.Header.Set("Content-Type", ct+"; charset=utf-8")
			return
		}
		// Do not modify binary types like image/*
	}
}
