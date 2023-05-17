BEGIN;

CREATE TABLE IF NOT EXISTS 
USERS
(
    uid serial primary key,
    login varchar unique,
    password_hash varchar, 
    current_balance real,
    withdrawn real,
    created_at timestamp default current_timestamp
);

CREATE TABLE IF NOT EXISTS 
ORDERS
(
    uid integer references users(uid),
    id bigint primary key,
    status varchar,
    accrual real,
    uploaded_at timestamp default current_timestamp
);

CREATE TABLE IF NOT EXISTS 
WITHDRAWALS
(
    uid integer references users(uid),
    id bigint primary key,
    sum real,
    processed_at timestamp default current_timestamp
);

COMMIT;
