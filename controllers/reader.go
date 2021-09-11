package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HttpRead struct {
	path string
}

func (read *HttpRead) Dial(c *gin.Context) {
	response, err := http.Get(read.path)
	if err != nil || response.StatusCode != http.StatusOK {
		c.Status(http.StatusServiceUnavailable)
		return
	}
	reader := response.Body
	contentLength := response.ContentLength
	contentType := response.Header.Get("Content-Type")

	extraHeaders := map[string]string{
		"Content-Disposition": `attachment; filename="test.png"`,
	}

	c.DataFromReader(http.StatusOK, contentLength, contentType, reader, extraHeaders)
}
