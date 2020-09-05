CREATE TABLE chat_messages(
   id SERIAL PRIMARY KEY,
   user_id INT NOT NULL,
   channel_id INT NOT NULL,
   content TEXT,
   sent_at TIMESTAMPTZ NOT NULL
);

ALTER TABLE chat_messages
 ADD CONSTRAINT fk_chat_messages_channel_id
 FOREIGN KEY (channel_id)
 REFERENCES channels (id);

 ALTER TABLE chat_messages
 ADD CONSTRAINT fk_chat_messages_user_id
 FOREIGN KEY (user_id)
 REFERENCES users (id);

CREATE INDEX idx_chat_message_sent_at ON chat_messages (sent_at);
CREATE INDEX idx_chat_message_user_id ON chat_messages (user_id);
CREATE INDEX idx_chat_message_channel_id ON chat_messages (channel_id);