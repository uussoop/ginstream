package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uussoop/ginstream"
)

func main() {
	r := gin.Default()

	request := "request"
	response := "response"
	// this can be .data.message
	path := "message"

	HandlerConf := ginstream.GeneralPurposeHandlerType{
		StreamHandlerFunc:    ginstream.SampleHandler,
		NonStreamHandlerFunc: ginstream.SampleNonstreamHandler,

		Timeout:           100 * time.Second,
		InputName:         &request,
		OutputName:        &response,
		StreamMessagePath: &path,
	}
	r.POST("/chat", ginstream.GeneralPurposeHandler(HandlerConf))

	r.Run(":8080")
}
