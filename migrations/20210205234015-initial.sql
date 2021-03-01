-- +migrate Up
CREATE TABLE IF NOT EXISTS market_user
(
    id            SERIAL PRIMARY KEY,
    first_name    VARCHAR(100) NOT NULL,
    last_name     VARCHAR(100) NOT NULL,
    national_code VARCHAR(10)  NOT NULL UNIQUE,
    password      VARCHAR(97)  NOT NULL,

    last_login    TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at    TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS symbol
(
    id            SERIAL PRIMARY KEY,
    isin          VARCHAR(10) NOT NULL UNIQUE,
    initial_price MONEY       NOT NULL,

    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TYPE ORDER_SIDE AS ENUM ('buy', 'sell');

CREATE TABLE IF NOT EXISTS market_order
(
    id              SERIAL PRIMARY KEY,
    user_id         INTEGER    NOT NULL REFERENCES market_user (id),
    symbol_id       INTEGER    NOT NULL REFERENCES symbol (id),
    quantity        INTEGER    NOT NULL,
    filled_quantity INTEGER                  DEFAULT 0,
    price           MONEY      NOT NULL,
    side            ORDER_SIDE NOT NULL,

    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    filled_at       TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    cancelled_at    TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS transaction
(
    id         SERIAL PRIMARY KEY,
    seller_id  INTEGER                  NOT NULL REFERENCES market_user (id),
    buyer_id   INTEGER                  NOT NULL REFERENCES market_user (id),
    symbol_id  INTEGER                  NOT NULL REFERENCES symbol (id),
    quantity   INTEGER                  NOT NULL,
    price      MONEY                    NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- +migrate Down
DROP TABLE IF EXISTS transaction;
DROP TABLE IF EXISTS market_order;
DROP TYPE IF EXISTS ORDER_SIDE;
DROP TABLE IF EXISTS market_user;
DROP TABLE IF EXISTS symbol;
