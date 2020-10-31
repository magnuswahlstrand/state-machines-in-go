# Getting started with `sqlc`

`sqlc` is a command line tool to generate type-safe Go code, based on SQL queries and schemas.
In this article I cover how to set up a very useful application that insert animals (data about them not actual animals) to a database, and keeps a list of them forever!

### Prerequisites
For this article, I use the following tools

1. Install `sqlc`
    1. `brew install kyleconroy/sqlc/sqlc`
2. Install `docker` - used to run local PostgresSQL instance 
    1. https://docs.docker.com/docker-for-mac/   
3. Install `psql` - used to connect to the local PostgresSQL database 
   1. `brew install postgresql`

## Getting started

First we need to setup our local database. Add the following to ***schema.sql***.

***schema.sql***
```postgresql
CREATE TABLE animals (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT (NOW() AT TIME ZONE 'UTC')
);
```

This will be used also be used by `sqlc` to generate Go structs. For now, we will use it to bootstrap our local database.

Let's start our local PostgresSQL database.
```bash
docker run -p 5432:5432 -it --rm postgres
``` 

Next we connect to it using `psql` and run the commands in ***schema.sql***  
```bash
> psql -h localhost --username=postgres -c "\i schema.sql"
CREATE TABLE
``` 

### Queries
Now that we have set up our table, we can run some queries. Open a psql shell
```bash
psql -h localhost --username=postgres
```

```postgresql
postgres=# INSERT INTO animals (name, type) VALUES ('Fido', 'dog');
INSERT 0 1
postgres=# INSERT INTO animals (name, type) VALUES ('Spot', 'cat');
INSERT 0 1
postgres=# INSERT INTO animals (name, type) VALUES ('Magnus', 'racoon');
INSERT 0 1
postgres=# SELECT * FROM animals
postgres-# ;
 id |  name  |  type  |          created_at
----+--------+--------+-------------------------------
  1 | Fido   | dog    | 2020-10-25 16:17:22.641659+00
  2 | Spot   | cat    | 2020-10-25 16:17:32.089118+00
  3 | Magnus | racoon | 2020-10-25 16:17:41.17342+00
```

We have successfully written and read some data! For now, these are the queries we want for this simple application.

### Setup for `sqlc`
We want to turn the queries and schema into code. `sqlc` needs a configuration file with the following things
1) location of database schema - `schema.sql`
2) location of queries schema - `queries.sql`
3) package name for generated code - `data`
3) path of generated code - `data/`

Create the file **sqlc.yaml**

```yaml
version: "1"
packages:
  - name: "data"
    path: "data"
    schema: "schema.sql"
    queries: "queries.sql"
```

After this file, we need to create `queries.sql`.
```postgresql
-- name: ListAnimals :many
SELECT * FROM animals
ORDER BY created_at;

-- name: CreateAnimal :exec
INSERT INTO animals (name, type) VALUES ($1, $2)
```

The comments are extra notation for `sqlc`.  
* `name: ListAnimals :many` - Generate a method called `ListAnimals` that returns a slice of `Animal`
* `name: CreateAnimal :exec` - Generate a method called `CreateAnimal` that returns an error

#### Generate code

Now we can run 
```bash
sqlc generate
```

This generates the following two files of interest

***data/models.go***
```go
package data

import (
	"time"
)

type Animal struct {
	ID        int32
	Name      string
	Type      string
	CreatedAt time.Time
}
```

***data/queries.sql.go***
```go
package data

import (
	"context"
)

const createAnimal = `-- name: CreateAnimal :exec
INSERT INTO animals (name, type) VALUES ($1, $2)
`

type CreateAnimalParams struct {
	Name string
	Type string
}

func (q *Queries) CreateAnimal(ctx context.Context, arg CreateAnimalParams) error {
	_, err := q.db.ExecContext(ctx, createAnimal, arg.Name, arg.Type)
	return err
}

const listAnimals = `-- name: ListAnimals :many
SELECT id, name, type, created_at FROM animals
ORDER BY created_at
`

func (q *Queries) ListAnimals(ctx context.Context) ([]Animal, error) {
	rows, err := q.db.QueryContext(ctx, listAnimals)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Animal
	for rows.Next() {
		var i Animal
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Type,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
```

### Using in application
Since we already have our Postgres instance up and running with database table defined. We can now use this generated code directly our application.

***main.go***
```go
package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"math/rand"
	"time"

	"github.com/kyeett/sqlc-example/data"
	"log"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = ""
)

var sourceName = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, "")

func main() {
	// Seed randomizer
	rand.Seed(time.Now().UTC().UnixNano())

	db, err := sql.Open("postgres", sourceName)
	if err != nil {
		log.Fatalf("failed to open connection to DB: %v", err)
	}
	database := data.New(db)

	// Create animal
	err = database.CreateAnimal(context.Background(), data.CreateAnimalParams{
		Name: "Fido",
		Type: "dog",
	})
	if err != nil {
		log.Fatal("failed to create animal", err)
	}
	fmt.Println("* new animal with created")

	// List animals
	animals, err := database.ListAnimals(context.Background())
	if err != nil {
		log.Fatal("failed to list animals", err)
	}

	fmt.Println("* complete list of animals")
	for _, animal := range animals {
		fmt.Printf("-- %d\t%s\t%s\n", animal.ID, animal.Name, animal.Type)
	}
}
```
Finally, if we run the application, we get the following output
```bash
> go run main.go
* new animal with created
* complete list of animals
-- 1	Fido	dog
-- 2	Fido	dog
```

Success!

## Conclusion
`sqlc` is a neat, easy to use tool, which makes working with Postgres in Go fun!

***Some things I liked***
1. Easy to set up and use. Very straight forward syntax
2. Changes to the database layer tracked easily by changes to `schema.sql` and `queries.sql`
3. It forces you to focus on the SQL queries; what do you need to insert and read from the database? When that is decided, the models in the code can be generated, not vice-versa.

A few other topics that would be worth covering in the future:
* Delete and get methods
* Returning a single inserted model using `:one` and `RETURNING *`
* Using transactions with `sqlc`
* Handling migrations
* Handling custom types, such as `github.com/google/uuid`
