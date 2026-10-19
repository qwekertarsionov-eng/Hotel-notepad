CREATE TABLE IF NOT EXISTS guests (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT      NOT NULL,
    full_name   VARCHAR(255) NOT NULL,
    phone       VARCHAR(50)  NOT NULL,
    room_number VARCHAR(50)  NOT NULL,
    check_in    DATE         NOT NULL,
    check_out   DATE         NOT NULL,
    note        TEXT         NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT fk_guests_user
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT chk_guests_dates
        CHECK (check_out > check_in)
);

CREATE INDEX IF NOT EXISTS ix_guests_user_created_at ON guests (user_id, created_at DESC);
