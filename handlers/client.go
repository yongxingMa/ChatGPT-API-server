package handlers

import (
	"github.com/ChatGPT-Hackers/ChatGPT-API-server/app"
	"github.com/ChatGPT-Hackers/ChatGPT-API-server/e"
	"github.com/ChatGPT-Hackers/ChatGPT-API-server/gtp"
	"log"
	"net/http"
	"strings"
	"time"

	// Import local packages
	"github.com/ChatGPT-Hackers/ChatGPT-API-server/types"
	"github.com/ChatGPT-Hackers/ChatGPT-API-server/utils"

	"github.com/gin-gonic/gin"
)

// // # Client routes
func Client_register(c *gin.Context) {
	// Make websocket connection
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	// Generate connection id
	id := utils.GenerateId()
	// Send connection id
	err = ws.WriteJSON(types.Message{
		Id:      id,
		Message: "Connection id",
	})
	if err != nil {
		return
	}
	// Wait for client to send connection id
	for {
		// Read message
		var message types.Message
		err = ws.ReadJSON(&message)
		if err != nil {
			return
		}
		// Check if the message is the connection id
		if message.Id == id {
			break
		} else {
			// This is probably a reconnect
			// Check if the connection id is in the pool
			connection, ok := connectionPool.Get(message.Id)
			if ok {
				// Close the old connection
				connection.Ws.Close()
				// Remove the connection from the pool
				connectionPool.Delete(message.Id)
			}
			id = message.Id
			break
		}
	}
	// Add connection to the pool
	connection := &types.Connection{
		Id: id,
		Ws: ws,
		// Set last message time to the beginning of time
		LastMessageTime: time.Time{},
		Heartbeat:       time.Now(),
	}
	connectionPool.Set(connection)
	// Debug
	println("New connection:", connection.Id)
}

// // # ChatGPT method
func ChatGPT(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		form ChatBody
	)

	httpCode, errCode := app.BindAndValid(c, &form)
	if errCode != e.SUCCESS {
		appG.Response(httpCode, errCode, nil)
		return
	}

	requestText := strings.TrimSpace(form.Content)
	reply, err := gtp.Completions(requestText)
	if err != nil {
		log.Printf("gtp request error: %v \n", err)
		//return nil
	}

	appG.Response(http.StatusOK, e.SUCCESS, reply)
}

type ChatBody struct {
	//TagID         int    `form:"tag_id" valid:"Required;Min(1)"`
	//Title         string `form:"title" valid:"Required;MaxSize(100)"`
	//Desc          string `form:"desc" valid:"Required;MaxSize(255)"`
	Content string `form:"content" valid:"Required;MaxSize(65535)"`
	//CreatedBy     string `form:"created_by" valid:"Required;MaxSize(100)"`
	//CoverImageUrl string `form:"cover_image_url" valid:"Required;MaxSize(255)"`
	//State         int    `form:"state" valid:"Range(0,1)"`
}
