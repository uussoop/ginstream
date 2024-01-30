package ginstream

import (
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/stream", StreamHandler(sampleHandler, 100*time.Millisecond))

	r.Run(":8080")
}
