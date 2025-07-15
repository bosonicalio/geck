--  Copyright (c) 2025 Tesserical s.r.l. All rights reserved.
--
--  This source code is the property of Tesserical s.r.l. and is intended for internal use only.
--  Unauthorized copying, distribution, or disclosure of this code, in whole or in part, is strictly prohibited.
--
--  For internal development purposes only. Not for public release.
--
--  For inquiries, contact: legal@tesserical.com
--

-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS platform_users (
    user_id VARCHAR(48) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    image_url TEXT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS platform_users;
-- +goose StatementEnd
