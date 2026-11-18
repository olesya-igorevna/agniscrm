-- +goose Up
CREATE TABLE users (
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT        NOT NULL,
    email      TEXT        NOT NULL,
    role       TEXT        NOT NULL CHECK (role IN ('admin', 'head', 'manager')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX users_email_lower_idx ON users (lower(email));

CREATE TABLE companies (
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT        NOT NULL,
    website    TEXT        NOT NULL DEFAULT '',
    phone      TEXT        NOT NULL DEFAULT '',
    comment    TEXT        NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE stages (
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT        NOT NULL,
    position   INTEGER     NOT NULL CHECK (position >= 1),
    outcome    TEXT        NOT NULL DEFAULT 'open' CHECK (outcome IN ('open', 'won', 'lost')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX stages_name_lower_idx ON stages (lower(name));

CREATE TABLE contacts (
    id         BIGSERIAL PRIMARY KEY,
    first_name TEXT        NOT NULL,
    last_name  TEXT        NOT NULL DEFAULT '',
    phone      TEXT        NOT NULL DEFAULT '',
    email      TEXT        NOT NULL DEFAULT '',
    telegram   TEXT        NOT NULL DEFAULT '',
    company_id BIGINT      REFERENCES companies(id) ON DELETE RESTRICT,
    owner_id   BIGINT      NOT NULL DEFAULT 1 REFERENCES users(id) ON DELETE RESTRICT,
    tags       TEXT[]      NOT NULL DEFAULT '{}',
    comment    TEXT        NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE deals (
    id          BIGSERIAL PRIMARY KEY,
    title       TEXT        NOT NULL,
    amount      BIGINT      NOT NULL DEFAULT 0 CHECK (amount >= 0),
    contact_id  BIGINT      NOT NULL REFERENCES contacts(id) ON DELETE RESTRICT,
    company_id  BIGINT      REFERENCES companies(id) ON DELETE RESTRICT,
    stage_id    BIGINT      NOT NULL REFERENCES stages(id) ON DELETE RESTRICT,
    owner_id    BIGINT      NOT NULL DEFAULT 1 REFERENCES users(id) ON DELETE RESTRICT,
    description TEXT        NOT NULL DEFAULT '',
    closed_at   TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX deals_stage_id_idx   ON deals (stage_id);
CREATE INDEX deals_contact_id_idx ON deals (contact_id);

CREATE TABLE tasks (
    id           BIGSERIAL PRIMARY KEY,
    title        TEXT        NOT NULL,
    description  TEXT        NOT NULL DEFAULT '',
    due_at       TIMESTAMPTZ NOT NULL,
    deal_id      BIGINT      REFERENCES deals(id)    ON DELETE SET NULL,
    contact_id   BIGINT      REFERENCES contacts(id) ON DELETE SET NULL,
    assignee_id  BIGINT      NOT NULL DEFAULT 1 REFERENCES users(id) ON DELETE RESTRICT,
    status       TEXT        NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'done')),
    completed_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX tasks_assignee_id_idx ON tasks (assignee_id);
CREATE INDEX tasks_due_at_idx       ON tasks (due_at);

-- +goose Down
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS deals;
DROP TABLE IF EXISTS contacts;
DROP TABLE IF EXISTS stages;
DROP TABLE IF EXISTS companies;
DROP TABLE IF EXISTS users;
