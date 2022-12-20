package handlers

import (
	"encoding/json"
	"time"

	"github.com/ChatGPT-Hackers/go-server/types"
	"github.com/ChatGPT-Hackers/go-server/utils"
	"github.com/gin-gonic/gin"
)

// // # API routes
func API_ask(c *gin.Context) {
	// Get request
	var request types.ChatGptRequest
	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid request",
		})
		return
	}
	// If Id is not set, generate a new one
	if request.MessageId == "" {
		request.MessageId = utils.GenerateId()
	}
	// If parent id is not set, generate a new one
	if request.ParentId == "" {
		request.ParentId = utils.GenerateId()
	}
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Failed to convert request to json",
		})
		return
	}
	// Get connection with the lowest load
	var connection *types.Connection
	connectionPool.Mu.RLock()
	// Check number of connections
	if len(connectionPool.Connections) == 0 {
		c.JSON(503, gin.H{
			"error": "No available clients",
		})
		return
	}
	for _, conn := range connectionPool.Connections {
		if connection == nil || conn.LastMessageTime.Before(connection.LastMessageTime) {
			connection = conn
		}
	}
	connectionPool.Mu.RUnlock()
	// Do not send request if connection currently has a request
	if connection.LastMessageTime.After(connection.Heartbeat) {
		c.JSON(503, gin.H{
			"error": "No available clients",
		})
		return
	}
	// Ping before sending request
	if !ping(connection.Id) {
		c.JSON(503, gin.H{
			"error": "Ping failed",
		})
		return
	}
	message := types.Message{
		Id:      utils.GenerateId(),
		Message: "ChatGptRequest",
		// Convert request to json
		Data: string(jsonRequest),
	}
	err = connection.Ws.WriteJSON(message)
	// Set last message time
	connection.LastMessageTime = time.Now()
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Failed to send request to the client",
		})
		// Delete connection
		connectionPool.Delete(connection.Id)
		return
	}
	// Wait for response
	// Wait for response with a timeout
	timeout := time.After(60 * time.Second)
	for {
		select {
		case <-timeout:
			c.JSON(504, gin.H{
				"error": "Timed out waiting for response from the client",
			})
			return
		default:
			// Read message
			var receive types.Message
			err = connection.Ws.ReadJSON(&receive)
			if err != nil {
				c.JSON(500, gin.H{
					"error": "Failed to read response from the client",
					"err":   err.Error(),
				})
				// Delete connection
				connectionPool.Delete(connection.Id)
				return
			}
			// Check if the message is the response
			if receive.Id == message.Id {
				// Convert response to ChatGptResponse
				var response types.ChatGptResponse
				err = json.Unmarshal([]byte(receive.Data), &response)
				if err != nil {
					c.JSON(500, gin.H{
						"error":    "Failed to convert response to ChatGptResponse",
						"response": receive,
					})
					return
				}
				// Send response
				c.JSON(200, response)
				// Heartbeat
				connection.Heartbeat = time.Now()
				return
			} else {
				// Error
				c.JSON(500, gin.H{
					"error": "Failed to find response from the client",
				})
				return
			}
		}
	}

}

func API_getConnections(c *gin.Context) {
	// Get connections
	var connections []*types.Connection
	connectionPool.Mu.RLock()
	for _, connection := range connectionPool.Connections {
		connections = append(connections, connection)
	}
	connectionPool.Mu.RUnlock()
	// Send connections
	c.JSON(200, gin.H{
		"connections": connections,
	})
}

func ping(connection_id string) bool {
	// Get connection
	connection, ok := connectionPool.Get(connection_id)
	// Send "ping" to the connection
	if ok {
		send := types.Message{
			Id:      utils.GenerateId(),
			Message: "ping",
		}
		err := connection.Ws.WriteJSON(send)
		if err != nil {
			// Delete connection
			connectionPool.Delete(connection_id)
			return false
		}
		// Wait for response with a timeout
		timeout := time.After(5 * time.Second)
		for {
			select {
			case <-timeout:
				return false
			default:
				// Read message
				var receive types.Message
				err = connection.Ws.ReadJSON(&receive)
				if err != nil {
					// Delete connection
					connectionPool.Delete(connection_id)
					return false
				}
				// Check if the message is the response
				if receive.Id == send.Id {
					return true
				}
			}
		}
	}
	return false
}
