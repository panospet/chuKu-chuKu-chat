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

type AppDb struct {
	Conn *sqlx.DB
}

func NewDb(dbPath string) (*AppDb, error) {
	db, err := sqlx.Connect("postgres", dbPath)
	if err != nil {
		return &AppDb{}, err
	}
	return &AppDb{Conn: db}, nil
}

func (o *AppDb) AddChannel(ch model.Channel) error {
	q := `insert into channels (name,creator,description,is_private) values($1,$2,$3,$4)`
	_, err := o.Conn.Exec(q, ch.Name, ch.Creator, ch.Description, ch.IsPrivate)
	if err != nil {
		return err
	}
	return nil
}

func (o *AppDb) ChannelLastMessages(channelName string, amount int) ([]model.Msg, error) {
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
	return messages, nil
}
func (o *AppDb) GetChannels() ([]model.Channel, error) {
	return nil, errors.New("not implemented")
}
func (o *AppDb) DeleteChannel(name string) error {
	return errors.New("not implemented")
}

func (o *AppDb) GetChannel(name string) (model.Channel, error) {
	q := `select id,name,creator,description,is_private,created_at from channels where name=$1`
	row := o.Conn.QueryRowx(q, name)
	var c model.Channel
	err := row.StructScan(&c)
	if err != nil {
		return model.Channel{}, errors.New(fmt.Sprintf("error getting channel: %s", err))
	}
	return c, nil
}

func (o *AppDb) AddUser(user model.User) error {
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

func (o *AppDb) GetUser(name string) (model.User, error) {
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

func (o *AppDb) GetUsers() ([]model.User, error) {
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

func (o *AppDb) RemoveUser(username string) error {
	if _, err := o.Conn.Exec("delete from users where name=$1", username); err != nil {
		return err
	}
	return nil
}

func (o *AppDb) AddSubscription(username string, channelName string) error {
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
	return nil
}

func (o *AppDb) AddMessage(m model.Msg) error {
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
