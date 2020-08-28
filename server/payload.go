package server

import (
	"github.com/gin-gonic/gin"
)

/*
Reference:  https://github.com/fanjq99/dnslog.git
*/
func (h *WebServer) Index(c *gin.Context) {
	body := `<h1>d4f800167a6e317f35454ed9024ebd420</h1>">`
	c.Data(200, "text/html; charset=utf-8", []byte(body))
}

func (h *WebServer) Status(c *gin.Context) {
	c.Data(200, "text/html; charset=utf-8", []byte(""))
}

func (h *WebServer) phpRFI(c *gin.Context) {
	body := `<?php  echo md5("GODNSLOG");` //694ef536e5d0245f203a1bcf8cbf3294
	c.Data(200, "text/html; charset=utf-8", []byte(body))
}

func (h *WebServer) xss(c *gin.Context) {
	body := `<script>prompt(98589956)</script>`
	c.Data(200, "text/html; charset=utf-8", []byte(body))
}
