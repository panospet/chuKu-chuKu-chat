# ChuKu-ChuKu-Chat

Ridiculous name for a chat application, I know.

### How to run

```
cd cmd/app
go run main.go
```

Then check if everything is running successfully with:

```
curl -XGET "http://localhost:8000/health"
```
if response is 200, you're good to go.

### Endpoints available

`GET /health`

`GET /channels`

`POST /channels` with body: 

```
{
  "name": "test channel", 
  "description": "test discussion"  
  "creator": "user1"
}
```

`DELETE /channels/{string}`

`GET /channels/{string}/lastMessages?amount={int}`

`GET /chat?username={string}`

