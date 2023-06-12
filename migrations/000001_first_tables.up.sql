CREATE TABLE IF NOT EXISTS person
(
    id       SERIAL NOT NULL PRIMARY KEY,
    name     TEXT   NOT NULL,
    password TEXT   NOT NULL
);

CREATE TABLE IF NOT EXISTS "group"
(
    id              SERIAL NOT NULL PRIMARY KEY,
    name            TEXT   NOT NULL,
    owner_id        INTEGER REFERENCES person (id),
    invitation_code TEXT   NOT NULL,
    balance         bigint NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS group_person
(
    id        SERIAL NOT NULL PRIMARY KEY,
    group_id  INTEGER REFERENCES "group" (id),
    person_id INTEGER REFERENCES person (id)
)