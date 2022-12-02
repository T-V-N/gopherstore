CREATE TABLE IF NOT EXISTS 
ORDERS
(
    uid integer references users(uid),
    id bigint primary key,
    status varchar,
    accrual real,
    uploaded_at timestamp default current_timestamp
);
