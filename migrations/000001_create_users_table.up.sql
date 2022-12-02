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
