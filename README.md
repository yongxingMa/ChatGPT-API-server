# ChatGPT API Server

# Quickstart 
## Setup
1. Install Go
2. `go install github.com/ChatGPT-Hackers/ChatGPT-API-server@latest`

# Build
1. `git clone https://github.com/ChatGPT-Hackers/ChatGPT-API-server/`
2. `cd ChatGPT-API-server`
3. `go build .`

# Usgae
`ChatGPT-API-server <port> <secret key>`

# Connect agents
Take note of your IP address or domain name. This could be `localhost` or a remote IP address. The default port is `8080`

Check out our [firefox agent](https://github.com/ChatGPT-Hackers/ChatGPT-API-agent). More versions in the works.

# Usage
```bash
 $ curl "http://localhost:8080/api/ask" -X POST --header 'Authorization: <API_KEY>' -d '{"content": "Hello world", "conversation_id": "<optional>", "parent_id": "<optional>"}'
 ```
