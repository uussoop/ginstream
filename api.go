package ginstream

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type GeneralPurposeHandlerType struct {
	// this handler function is to handle streams any stream must be pushed into event and message channels and when you are done you should either close the channels or push the done channels
	StreamHandlerFunc func(MessageChannel *chan any, EventNameChannel *chan string, DoneChannel *chan bool, Input *string)
	// this handler function is to handle NONstreams the response should be returned in this scenario
	NonStreamHandlerFunc func(*string) any
	//this is the timeout duration which only works on streams
	Timeout time.Duration
	// this is the field that you will assign input data's name in gin context to be proccesed in handlers
	InputName *string
	// this is the field that you will assign output data's name in gin context to be proccesed in handlers
	OutputName *string
	// this is used to record the messages in stream if needed if empty the whole message is reacorded
	StreamMessagePath *string
}
type streamType struct {
	stream bool
}

func structToMap(p any) map[string]any {
	val := reflect.ValueOf(p)
	typ := val.Type()

	m := make(map[string]any)

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i).Interface()

		m[strings.ToLower(field.Name)] = value
	}
	return m
}
func parseStreamPath(input any, path *string) *string {
	pathParsed := strings.Split(strings.ToLower(*path), ".")

	message := ""
	var lstInput any
	lstInput = input
	for index, v := range pathParsed {

		t := reflect.TypeOf(lstInput)

		switch t.Kind() {
		case reflect.Slice:
			valuelist := reflect.ValueOf(lstInput)
			in, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				message += fmt.Sprintf("%v", lstInput)
				return &message
			}
			if uint(in) <= uint(valuelist.Len())-1 {
				if len(pathParsed)-1 == index {
					message += fmt.Sprintf("%v", valuelist.Index(int(in)).Interface())
					return &message
				}
			} else {
				fmt.Println("err")
				message += fmt.Sprintf("%v", lstInput)
				return &message
			}
		case reflect.Map:
			// fmt.Println(t.Key())
			// switch t.Key().Kind() {
			// case reflect.String:

			// 	value ,ok := input.(map[string]int)
			// 	if !ok {

			// 	}
			// 	val, ok := value[v]
			// 	if !ok {
			// 		message += fmt.Sprintf("%v", value)
			// 		return &message
			// 	}
			// 	if len(pathParsed)-1 == index {

			// 		message += fmt.Sprintf("%v", val)
			// 		return &message
			// 	}
			// 	m := parseStreamPath(val, &v)
			// 	message += *m
			// default:

			// }
			panic("only accepting structs or basic types")
		case reflect.Struct:

			value := structToMap(lstInput)

			val, ok := value[v]
			if !ok {
				message += fmt.Sprintf("%v", value)
				return &message
			}
			if len(pathParsed)-1 == index {

				message += fmt.Sprintf("%v", val)
				return &message
			}
			lstInput = val
			// m := parseStreamPath(val, &v)

		default:
			if index > 0 {
				continue
			}
			message += fmt.Sprintf("%v", input)
			break
		}
	}
	if message == "" {
		message = fmt.Sprintf("%v", input)
	}
	return &message
}
func SampleHandler(
	messageChannel *chan any,
	eventNameChannel *chan string,
	donechannel *chan bool,
	Input *string,
) {
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
		time.Sleep(1 * time.Second)
	}
	// Close the channel when the messages are sent
	close(*donechannel)
}
func SampleNonstreamHandler(
	Input *string,
) any {
	return struct {
		Message string
		Status  int
	}{
		Message: "hi",
		Status:  123,
	}
}

var DefaultRequestInput = "request"
var DefaultResponseInput = "response"

func GeneralPurposeHandler(
	g GeneralPurposeHandlerType,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		isStream := false
		var s streamType
		c.ShouldBindJSON(s)
		if c.GetHeader("Content-Type") == "text/event-stream" || s.stream {
			isStream = true
		}
		var req *string

		if *g.InputName == "" || g.InputName == nil {
			g.InputName = &DefaultRequestInput
		}
		if *g.OutputName == "" || g.OutputName == nil {
			g.OutputName = &DefaultResponseInput

		}
		requ, ok := c.Get(*g.InputName)
		req, ok = requ.(*string)

		if !ok {
			c.JSON(http.StatusBadRequest, "bad request")
			return
			// req = &DefaultRequestInput
		}
		if g.NonStreamHandlerFunc == nil {
			c.JSON(http.StatusBadRequest, "non stream config not set")
			return
		}
		if g.StreamHandlerFunc == nil {
			isStream = false
		}
		switch isStream {
		case true:
			streamHandler(
				g.StreamHandlerFunc,

				g.Timeout,
				c,
				req,
				g.OutputName,
				g.StreamMessagePath,
			)
		case false:
			nonStreamHandler(

				g.NonStreamHandlerFunc,
				g.Timeout,
				c,
				req,
				g.OutputName,
				g.StreamMessagePath,
			)
		default:
			nonStreamHandler(

				g.NonStreamHandlerFunc,
				g.Timeout,
				c,
				req,
				g.OutputName,
				g.StreamMessagePath,
			)

		}
	}

}

func streamHandler(
	HandlerFunc func(MessageChannel *chan any, EventNameChannel *chan string, DoneChannel *chan bool, Input *string),
	timeout time.Duration,
	c *gin.Context,
	input *string,
	outputName *string,
	streamPath *string,
) {

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	messageChannel := make(chan any)
	eventNameChannel := make(chan string)
	donechannel := make(chan bool)

	go HandlerFunc(&messageChannel, &eventNameChannel, &donechannel, input)
	if streamPath == nil {
		*streamPath = "."
	}
	fullmessage := ""
	for {
		select {
		case message, ok := <-messageChannel:

			if !ok {
				return
			}
			m := parseStreamPath(message, streamPath)
			fullmessage += " " + *m
			select {
			case event, ok := <-eventNameChannel:
				if !ok {
					return
				}
				c.SSEvent(event, message)
				c.Writer.Flush()
			case <-c.Request.Context().Done():

				return
			case done, ok := <-donechannel:
				if !ok {
					close(messageChannel)
					close(eventNameChannel)
					c.Set(*outputName, fullmessage)

					return
				}
				if done {
					close(messageChannel)
					close(eventNameChannel)
					close(donechannel)
					c.Set(*outputName, fullmessage)
					return
				}
			case <-time.After(timeout):

				return

			}

		case <-c.Request.Context().Done():

			return
		case done, ok := <-donechannel:
			if !ok {
				close(messageChannel)
				close(eventNameChannel)
				c.Set(*outputName, fullmessage)
				return
			}
			if done {
				close(messageChannel)
				close(eventNameChannel)
				close(donechannel)
				c.Set(*outputName, fullmessage)
				return
			}
		case <-time.After(timeout):

			return

		}
	}
}

func nonStreamHandler(
	HandlerFunc func(*string) any,
	timeout time.Duration,
	c *gin.Context,
	input *string,
	outputName *string,
	streamPath *string,
) {
	if streamPath == nil {
		*streamPath = "."
	}
	resp := HandlerFunc(input)
	c.Set(*outputName, parseStreamPath(resp, streamPath))
	c.JSON(http.StatusOK, resp)
	return
}
