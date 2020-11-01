-- name: GetOrder :one
SELECT *
FROM orders
WHERE id = ($1);

-- name: CreateOrder :one
INSERT INTO orders (state)
VALUES ($1)
RETURNING *;

-- name: UpdateOrderState :exec
UPDATE orders
SET state = ($1)
WHERE id = ($2);