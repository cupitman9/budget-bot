CREATE TABLE users
(
    chat_id    bigint       NOT NULL PRIMARY KEY,
    username   varchar(255) NOT NULL DEFAULT '',
    language   varchar(50)  NOT NULL DEFAULT '',
    created_at timestamp    NOT NULL DEFAULT now()
);

CREATE TABLE categories
(
    id         bigserial PRIMARY KEY,
    name       varchar(255) NOT NULL,
    chat_id    bigint       NOT NULL REFERENCES users (chat_id),
    created_at timestamp    NOT NULL DEFAULT now()
);

CREATE TABLE transactions
(
    chat_id          bigint                  NOT NULL REFERENCES users (chat_id),
    category_id      bigint    DEFAULT 0     NOT NULL REFERENCES categories (id),
    amount           numeric(10, 2)          NOT NULL,
    created_at       timestamp DEFAULT now() NOT NULL,
    transaction_type smallint                NOT NULL, -- 1 = income 2 = expense
    PRIMARY KEY (chat_id, category_id, transaction_type, created_at)
);
