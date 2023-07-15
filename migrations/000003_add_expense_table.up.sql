CREATE TABLE expense(
    id SERIAL PRIMARY KEY,
    amount_in_cents INT,
    person_id INT REFERENCES person (id),
    group_id INT REFERENCES "group" (id)
)