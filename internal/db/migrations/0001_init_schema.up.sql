CREATE TABLE categories (
	id         BIGSERIAL PRIMARY KEY,
	slug       TEXT NOT NULL UNIQUE,
	name       TEXT NOT NULL,
	sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE products (
	id           BIGSERIAL PRIMARY KEY,
	slug         TEXT NOT NULL UNIQUE,
	category_id  BIGINT NOT NULL REFERENCES categories(id),
	name         TEXT NOT NULL,
	description  TEXT NOT NULL DEFAULT '',
	age_label    TEXT NOT NULL DEFAULT '',
	icon         TEXT NOT NULL DEFAULT 'blocks',
	accent_color TEXT NOT NULL DEFAULT 'teal',
	price_cents  BIGINT NOT NULL,
	currency     TEXT NOT NULL DEFAULT 'UAH',
	stock_qty    INTEGER NOT NULL DEFAULT 0,
	is_active    BOOLEAN NOT NULL DEFAULT true,
	created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_products_category ON products(category_id);

CREATE TABLE orders (
	id               BIGSERIAL PRIMARY KEY,
	customer_name    TEXT NOT NULL,
	phone            TEXT NOT NULL,
	email            TEXT NOT NULL DEFAULT '',
	city             TEXT NOT NULL DEFAULT '',
	address          TEXT NOT NULL DEFAULT '',
	comment          TEXT NOT NULL DEFAULT '',
	status           TEXT NOT NULL DEFAULT 'new',
	payment_status   TEXT NOT NULL DEFAULT 'pending',
	payment_provider TEXT NOT NULL DEFAULT '',
	total_cents      BIGINT NOT NULL DEFAULT 0,
	currency         TEXT NOT NULL DEFAULT 'UAH',
	created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE order_items (
	id               BIGSERIAL PRIMARY KEY,
	order_id         BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
	product_id       BIGINT NOT NULL REFERENCES products(id),
	product_name     TEXT NOT NULL,
	unit_price_cents BIGINT NOT NULL,
	quantity         INTEGER NOT NULL
);
CREATE INDEX idx_order_items_order ON order_items(order_id);

CREATE TABLE admin_users (
	id            BIGSERIAL PRIMARY KEY,
	email         TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL
);
