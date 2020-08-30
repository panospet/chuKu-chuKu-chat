# ChuKu-ChuKu-Chat

Ridiculous name for a chat application, I know.

```
	     _______________                        |*\_/*|________
	    |  ___________  |     .-.     .-.      ||_/-\_|______  |
	    | |           | |    .****. .****.     | |           | |
	    | |   0   0   | |    .*****.*****.     | |   0   0   | |
	    | |     -     | |     .**chuku**.      | |     -     | |
	    | |   \___/   | |      .*chuku*.       | |   \___/   | |
	    | |___     ___| |       .*****.        | |___________| |
	    |_____|\_/|_____|        .***.         |_______________|
	      _|__|/ \|_|_.............*.............._|________|_
	     / ********** \                          / ********** \
	   /  ****chu*****  \                      /  *****ku*****  \
	  --------------------                    --------------------
```

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

`POST /channels/subscription` with body: 

```
{
  "channel": "channel name"  
  "user": "testuser"
}
```

`DELETE /channels/{channelName}`

`GET /channels/{string}/lastMessages?amount={int}`

`GET /chat?username={string}` for creating a user && a websockeet connection

`GET /users`

`GET /users/{username}`

`GET /users/{username}/channels`

