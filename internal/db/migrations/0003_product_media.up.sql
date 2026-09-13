CREATE TABLE product_media (
	id         BIGSERIAL PRIMARY KEY,
	product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
	media_type TEXT NOT NULL CHECK (media_type IN ('photo', 'video')),
	url        TEXT NOT NULL,
	sort_order INTEGER NOT NULL DEFAULT 0,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_product_media_product ON product_media(product_id, sort_order);
