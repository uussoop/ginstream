package ginstream

import (
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	request := "request"
	response := "response"

	HandlerConf := GeneralPurposeHandlerType{
		StreamHandlerFunc:    sampleHandler,
		NonStreamHandlerFunc: sampleNonstreamHandler,

		Timeout:    100 * time.Millisecond,
		InputName:  &request,
		OutputName: &response,
	}
	r.GET("/chat", GeneralPurposeHandler(HandlerConf))

	r.Run(":8080")
}
