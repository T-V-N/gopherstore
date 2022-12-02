CREATE TABLE IF NOT EXISTS 
WITHDRAWALS
(
    uid integer references users(uid),
    id bigint primary key,
    sum real,
    processed_at timestamp default current_timestamp
);