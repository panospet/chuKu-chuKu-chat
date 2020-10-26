CREATE TABLE channels(
   id SERIAL PRIMARY KEY,
   name VARCHAR(512) UNIQUE NOT NULL,
   creator VARCHAR(512) NOT NULL,
   description TEXT default '',
   is_private BOOLEAN NOT NULL,
   created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_channel_name ON channels (name);

insert into channels (name,creator,description,is_private,created_at) values('general','KickIt-bot','general discussion','f',now());