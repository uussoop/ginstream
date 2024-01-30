package ginstream

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func sampleHandler(messageChannel *chan any, eventNameChannel *chan string) {
	for i := 0; i < 5; i++ {
		// message := fmt.Sprintf("Message %d", i+1)
		message := struct {
			Message string
			Count   int
		}{
			Message: fmt.Sprintf("Message %d", i+1),
			Count:   i + 1,
		}
		// Send the message to the client
		*messageChannel <- message
		*eventNameChannel <- "message"

		// Introduce a delay to simulate some processing
		time.Sleep(1 * time.Nanosecond)
	}
	// Close the channel when the messages are sent
	close(*messageChannel)
}

func StreamHandler(
	HandlerFunc func(MessageChannel *chan any, EventNameChannel *chan string),
	timeout time.Duration,
) gin.HandlerFunc {

	return func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		messageChannel := make(chan any)
		eventNameChannel := make(chan string)

		defer close(messageChannel)
		defer close(eventNameChannel)

		go HandlerFunc(&messageChannel, &eventNameChannel)

		for {
			select {
			case message, ok := <-messageChannel:

				if !ok {
					return
				}
				select {
				case event, ok := <-eventNameChannel:
					if !ok {
						return
					}
					c.SSEvent(event, message)
					c.Writer.Flush()
				case <-c.Request.Context().Done():

					return
				case <-time.After(timeout):

					return

				}

			case <-c.Request.Context().Done():

				return
			case <-time.After(timeout):

				return

			}
		}
	}

}
