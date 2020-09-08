CREATE TABLE user_to_channel(
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    channel_id INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE user_to_channel
 ADD CONSTRAINT fk_user_to_channel_user_id
 FOREIGN KEY (user_id)
 REFERENCES users (id)
 ON DELETE CASCADE;

ALTER TABLE user_to_channel
 ADD CONSTRAINT fk_user_to_channel_channel_id
 FOREIGN KEY (channel_id)
 REFERENCES channels (id)
 ON DELETE CASCADE;