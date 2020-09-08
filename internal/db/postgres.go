package db

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"chuKu-chuKu-chat/internal/model"
)

type PostgresDb struct {
	Conn *sqlx.DB
	rdb  *redis.Client
}

func NewPostgresDb(dbPath string, rdb *redis.Client) (*PostgresDb, error) {
	db, err := sqlx.Connect("postgres", dbPath)
	if err != nil {
		return &PostgresDb{}, err
	}
	return &PostgresDb{Conn: db, rdb: rdb}, nil
}

func (o *PostgresDb) AddChannel(ch model.Channel) error {
	user, err := o.GetUser(ch.Creator)
	if err != nil {
		return errors.New("user does not exist")
	}

	q := `insert into channels (name,creator,description,is_private) values($1,$2,$3,$4)`
	_, err = o.Conn.Exec(q, ch.Name, ch.Creator, ch.Description, ch.IsPrivate)
	if err != nil {
		return err
	}
	user.AddChannel(ch.Name)
	return user.RefreshChannels(o.rdb)
}

func (o *PostgresDb) ChannelLastMessages(channelName string, amount int) ([]model.Msg, error) {
	c, err := o.GetChannel(channelName)
	if err != nil {
		return nil, err
	}
	var messages []model.Msg
	q := `select chat_messages.id,users.name,content,sent_at from chat_messages left join users on 
users.id=chat_messages.user_id where channel_id=$1 order by sent_at desc limit $2;`
	rows, err := o.Conn.Query(q, c.Id, amount)
	for rows.Next() {
		var id string
		var userName string
		var content string
		var sentAt time.Time
		if err := rows.Scan(&id, &userName, &content, &sentAt); err != nil {
			return nil, err
		}
		m := model.Msg{
			Id:        id,
			Content:   content,
			Channel:   c.Name,
			User:      userName,
			Timestamp: sentAt,
		}
		messages = append(messages, m)
	}
	return reverseMessages(messages), nil
}

func reverseMessages(messages []model.Msg) []model.Msg {
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages
}

func (o *PostgresDb) GetChannels() ([]model.Channel, error) {
	q := `select id,name,creator,description,is_private,created_at from channels`
	rows, err := o.Conn.Queryx(q)
	if err != nil {
		fmt.Println("error while trying query", err)
		return nil, err
	}
	var out []model.Channel
	for rows.Next() {
		var c model.Channel
		if err := rows.StructScan(&c); err != nil {
			fmt.Println("error while scanning", err)
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func (o *PostgresDb) DeleteChannel(name string) error {
	q := "delete from channels where name=$1"
	if _, err := o.Conn.Exec(q, name); err != nil {
		fmt.Println("error executing delete", err)
		return err
	}
	return nil
}

func (o *PostgresDb) GetChannel(name string) (model.Channel, error) {
	q := `select id,name,creator,description,is_private,created_at from channels where name=$1`
	row := o.Conn.QueryRowx(q, name)
	var c model.Channel
	err := row.StructScan(&c)
	if err != nil {
		return model.Channel{}, errors.New(fmt.Sprintf("error getting channel: %s", err))
	}
	return c, nil
}

func (o *PostgresDb) AddUser(user model.User) error {
	tx, err := o.Conn.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var userId int
	q := `insert into users(name) values($1) returning id`
	err = tx.Get(&userId, q, user.Username)
	if err != nil {
		return err
	}
	query, args, err := sqlx.In("SELECT id FROM channels WHERE name IN (?);", user.GetChannels())
	query = tx.Rebind(query)
	var ids []int
	rows, err := tx.Query(query, args...)
	if err != nil {
		fmt.Println("error in get")
		return err
	}
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return err
		}
		ids = append(ids, id)
	}

	for _, id := range ids {
		q = `insert into user_to_channel (user_id,channel_id) values ($1,$2) on conflict do nothing`
		if _, err := tx.Exec(q, userId, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (o *PostgresDb) GetUser(name string) (model.User, error) {
	q := `select u.id,u.name,u.created_at, array(select name from channels ch join user_to_channel utc on 
utc.channel_id=ch.id and utc.user_id=u.id) from users u where u.name=$1`
	row := o.Conn.QueryRow(q, name)
	var userId int
	var userName string
	var userCreatedAt time.Time
	var channels []string
	if err := row.Scan(&userId, &userName, &userCreatedAt, pq.Array(&channels)); err != nil {
		return model.User{}, err
	}
	u := model.User{
		Id:               userId,
		Username:         userName,
		StopListenerChan: make(chan struct{}),
		MessageChan:      make(chan redis.Message),
		CreatedAt:        userCreatedAt,
		Channels:         channels,
	}
	return u, nil
}

func (o *PostgresDb) GetUsers() ([]model.User, error) {
	q := `select u.id,u.name,u.created_at, array(select name from channels ch join user_to_channel utc on 
utc.channel_id=ch.id and utc.user_id=u.id) from users u`

	rows, err := o.Conn.Query(q)
	if err != nil {
		return []model.User{}, err
	}
	var users []model.User
	for rows.Next() {
		var userId int
		var userName string
		var userCreatedAt time.Time
		var channels []string
		if err := rows.Scan(&userId, &userName, &userCreatedAt, pq.Array(&channels)); err != nil {
			return []model.User{}, err
		}
		u := model.User{
			Id:               userId,
			Username:         userName,
			StopListenerChan: make(chan struct{}),
			MessageChan:      make(chan redis.Message),
			CreatedAt:        userCreatedAt,
			Channels:         channels,
		}
		users = append(users, u)
	}
	return users, nil
}

func (o *PostgresDb) RemoveUser(username string) error {
	if _, err := o.Conn.Exec("delete from users where name=$1", username); err != nil {
		return err
	}
	return nil
}

func (o *PostgresDb) AddSubscription(username string, channelName string) error {
	u, err := o.GetUser(username)
	if err != nil {
		return err
	}
	c, err := o.GetChannel(channelName)
	if err != nil {
		return err
	}
	q := `insert into user_to_channel (user_id,channel_id) values ($1,$2) on conflict do nothing`
	if _, err := o.Conn.Exec(q, u.Id, c.Id); err != nil {
		return err
	}
	u.AddChannel(channelName)
	return u.RefreshChannels(o.rdb)
}

func (o *PostgresDb) AddMessage(m model.Msg) error {
	u, err := o.GetUser(m.User)
	if err != nil {
		return err
	}
	c, err := o.GetChannel(m.Channel)
	if err != nil {
		return err
	}
	q := `insert into chat_messages (id,user_id,channel_id,content,sent_at) values ($1,$2,$3,$4,$5)`
	if _, err := o.Conn.Exec(q, m.Id, u.Id, c.Id, m.Content, m.Timestamp); err != nil {
		return errors.New(fmt.Sprintf("error storing message: %s", err))
	}
	return nil
}
