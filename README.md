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

Currently there are two modes.
- The "dummy" mode, which keeps everything in-memory. In this mode, no persistence will take place. After closing 
the application everything will vanish.
- The "db" mode, which keeps everything inside a PostgreSQL database.

### Choose between modes

To choose between modes, edit the file `config/config.yml`. The `mode` field can be either `"dummy"` or `"db"`.

#### Start Redis

Both modes need a Redis database. To start a Redis instance on your machine, type:
```
make redis-start
```

#### Start database (for "db" mode)

To start a PostgreSQL instance on your machine, and run all the migrations available, type:

```
make db-start
make migrate
```

You can login in the database anytime you want, by typing:
```
make db-login
```

#### Run the application!

Finally. You can run the application by typing:

```
make run
```
Et voila!

#### Healthcheck

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

`GET /chat?username={string}` for creating a user && a websocket connection

`GET /users`

`GET /users/{username}`

`DELETE /users/{username}`

`GET /users/{username}/channels`

