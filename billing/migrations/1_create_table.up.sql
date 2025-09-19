CREATE TYPE currency AS ENUM ('GEL', 'USD');

CREATE TABLE IF NOT EXISTS bills (
  id          SERIAL PRIMARY KEY,
  billing_id  TEXT UNIQUE,
  status      TEXT NOT NULL DEFAULT 'OPEN',
  currency    currency NOT NULL,
  total       BIGINT NOT NULL DEFAULT 0,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  closed_at   TIMESTAMPTZ
);


CREATE TABLE IF NOT EXISTS bill_items (
  id        SERIAL PRIMARY KEY,
  bill_id   TEXT NOT NULL REFERENCES bills(billing_id) ON DELETE CASCADE,
  name      TEXT NOT NULL,
  price     BIGINT NOT NULL,
  idemp_key TEXT NOT NULL,
  CONSTRAINT bill_item_unique UNIQUE (idemp_key)
);

CREATE TABLE IF NOT EXISTS bill_exchanges (
  id              SERIAL PRIMARY KEY,
  bill_id         TEXT NOT NULL REFERENCES bills(billing_id) ON DELETE CASCADE,
  base_currency   currency NOT NULL,
  target_currency currency NOT NULL,
  rate            DECIMAL(6,4) NOT NULL,
  total           BIGINT, -- this will be in the smallest unit of the `target_currency`
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
)
