-- name: GetOrder :one
SELECT *
FROM orders
WHERE id = ($1);

-- name: CreateOrder :one
INSERT INTO orders (name, state)
VALUES ($1, $2) RETURNING *;