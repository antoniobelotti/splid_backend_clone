CREATE TABLE transfer
(
    id              SERIAL PRIMARY KEY,
    amount_in_cents INT,
    sender_id       INT REFERENCES person (id),
    receiver_id     INT REFERENCES person (id),
    group_id        INT REFERENCES "group" (id)
)