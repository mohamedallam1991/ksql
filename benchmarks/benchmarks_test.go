package benchmarks

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/vingarcia/ksql"
	"github.com/vingarcia/ksql/kpgx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var UsersTable = ksql.NewTable("users")

func BenchmarkInsert(b *testing.B) {
	ctx := context.Background()

	driver := "postgres"
	connStr := "host=localhost port=5432 user=postgres password=postgres dbname=ksql sslmode=disable"

	type User struct {
		ID   int    `ksql:"id" db:"id"`
		Name string `ksql:"name" db:"name"`
		Age  int    `ksql:"age" db:"age"`
	}

	b.Run("ksql/sql-adapter", func(b *testing.B) {
		ksqlDB, err := ksql.New(driver, connStr, ksql.Config{
			MaxOpenConns: 1,
		})
		if err != nil {
			b.Fatalf("error creating ksql client: %s", err)
		}

		err = recreateTable(connStr)
		if err != nil {
			b.Fatalf("error creating table: %s", err.Error())
		}

		b.Run("insert-one", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				err := ksqlDB.Insert(ctx, UsersTable, &User{
					Name: strconv.Itoa(i),
					Age:  i,
				})
				if err != nil {
					b.Fatalf("insert error: %s", err.Error())
				}
			}
		})
	})

	b.Run("ksql/pgx-adapter", func(b *testing.B) {
		kpgxDB, err := kpgx.New(ctx, connStr, ksql.Config{
			MaxOpenConns: 1,
		})
		if err != nil {
			b.Fatalf("error creating kpgx client: %s", err)
		}

		err = recreateTable(connStr)
		if err != nil {
			b.Fatalf("error creating table: %s", err.Error())
		}

		b.Run("insert-one", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				err := kpgxDB.Insert(ctx, UsersTable, &User{
					Name: strconv.Itoa(i),
					Age:  i,
				})
				if err != nil {
					b.Fatalf("insert error: %s", err.Error())
				}
			}
		})
	})

	b.Run("sql", func(b *testing.B) {
		sqlDB, err := sql.Open(driver, connStr)
		if err != nil {
			b.Fatalf("error creating sql client: %s", err)
		}
		sqlDB.SetMaxOpenConns(1)

		err = recreateTable(connStr)
		if err != nil {
			b.Fatalf("error creating table: %s", err.Error())
		}

		b.Run("insert-one", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				user := User{
					Name: strconv.Itoa(i),
					Age:  i,
				}
				rows, err := sqlDB.Query(
					`INSERT INTO users(name, age) VALUES ($1, $2) RETURNING id`,
					user.Name, user.Age,
				)
				if err != nil {
					b.Fatalf("insert error: %s", err.Error())
				}
				if !rows.Next() {
					b.Fatalf("missing id from inserted record")
				}
				rows.Scan(&user.ID)
				err = rows.Close()
				if err != nil {
					b.Fatalf("error closing rows")
				}
			}
		})
	})

	b.Run("sqlx", func(b *testing.B) {
		sqlxDB, err := sqlx.Open(driver, connStr)
		if err != nil {
			b.Fatalf("error creating sqlx client: %s", err)
		}
		sqlxDB.SetMaxOpenConns(1)

		err = recreateTable(connStr)
		if err != nil {
			b.Fatalf("error creating table: %s", err.Error())
		}

		b.Run("insert-one", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				user := User{
					Name: strconv.Itoa(i),
					Age:  i,
				}
				rows, err := sqlxDB.NamedQuery(
					`INSERT INTO users(name, age) VALUES (:name, :age) RETURNING id`,
					user,
				)
				if err != nil {
					b.Fatalf("insert error: %s", err.Error())
				}
				if !rows.Next() {
					b.Fatalf("missing id from inserted record")
				}
				rows.Scan(&user.ID)
				err = rows.Close()
				if err != nil {
					b.Fatalf("error closing rows")
				}
			}
		})
	})

	b.Run("gorm-adapter", func(b *testing.B) {
		gormDB, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
		if err != nil {
			b.Fatalf("error creating gorm client: %s", err)
		}

		err = recreateTable(connStr)
		if err != nil {
			b.Fatalf("error creating table: %s", err)
		}

		b.Run("insert-one", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				err := gormDB.Table("users").Create(&User{
					Name: strconv.Itoa(i),
					Age:  i,
				}).Error
				if err != nil {
					b.Fatalf("insert error: %s", err)
				}
			}
		})
	})
}

func BenchmarkQuery(b *testing.B) {
	ctx := context.Background()

	driver := "postgres"
	connStr := "host=localhost port=5432 user=postgres password=postgres dbname=ksql sslmode=disable"

	type User struct {
		ID   int    `ksql:"id" db:"id"`
		Name string `ksql:"name" db:"name"`
		Age  int    `ksql:"age" db:"age"`
	}

	b.Run("ksql/sql-adapter", func(b *testing.B) {
		ksqlDB, err := ksql.New(driver, connStr, ksql.Config{
			MaxOpenConns: 1,
		})
		if err != nil {
			b.Fatalf("error creating ksql client: %s", err)
		}

		err = recreateTable(connStr)
		if err != nil {
			b.Fatalf("error creating table: %s", err.Error())
		}

		err = insertUsers(connStr, 100)
		if err != nil {
			b.Fatalf("error inserting users: %s", err.Error())
		}

		b.Run("single-row", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var user User
				err := ksqlDB.QueryOne(ctx, &user, `SELECT * FROM users OFFSET $1 LIMIT 1`, i%100)
				if err != nil {
					b.Fatalf("query error: %s", err.Error())
				}
			}
		})

		b.Run("multiple-rows", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var users []User
				err := ksqlDB.Query(ctx, &users, `SELECT * FROM users OFFSET $1 LIMIT 10`, i%90)
				if err != nil {
					b.Fatalf("query error: %s", err.Error())
				}
				if len(users) < 10 {
					b.Fatalf("expected 10 scanned users, but got: %d", len(users))
				}
			}
		})
	})

	b.Run("ksql/pgx-adapter", func(b *testing.B) {
		kpgxDB, err := kpgx.New(ctx, connStr, ksql.Config{
			MaxOpenConns: 1,
		})
		if err != nil {
			b.Fatalf("error creating kpgx client: %s", err)
		}

		err = recreateTable(connStr)
		if err != nil {
			b.Fatalf("error creating table: %s", err.Error())
		}

		err = insertUsers(connStr, 100)
		if err != nil {
			b.Fatalf("error inserting users: %s", err.Error())
		}

		b.Run("single-row", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var user User
				err := kpgxDB.QueryOne(ctx, &user, `SELECT * FROM users OFFSET $1 LIMIT 1`, i%100)
				if err != nil {
					b.Fatalf("query error: %s", err.Error())
				}
			}
		})

		b.Run("multiple-rows", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var users []User
				err := kpgxDB.Query(ctx, &users, `SELECT * FROM users OFFSET $1 LIMIT 10`, i%90)
				if err != nil {
					b.Fatalf("query error: %s", err.Error())
				}
				if len(users) < 10 {
					b.Fatalf("expected 10 scanned users, but got: %d", len(users))
				}
			}
		})
	})

	b.Run("sql", func(b *testing.B) {
		sqlDB, err := sqlx.Open(driver, connStr)
		if err != nil {
			b.Fatalf("error creating sql client: %s", err)
		}
		sqlDB.SetMaxOpenConns(1)

		err = recreateTable(connStr)
		if err != nil {
			b.Fatalf("error creating table: %s", err.Error())
		}

		err = insertUsers(connStr, 100)
		if err != nil {
			b.Fatalf("error inserting users: %s", err.Error())
		}

		b.Run("single-row", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var user User
				rows, err := sqlDB.QueryContext(ctx, `SELECT id, name, age FROM users OFFSET $1 LIMIT 1`, i%100)
				if err != nil {
					b.Fatalf("query error: %s", err.Error())
				}
				if !rows.Next() {
					b.Fatalf("missing user from inserted record, offset: %d", i%100)
				}
				err = rows.Scan(&user.ID, &user.Name, &user.Age)
				if err != nil {
					b.Fatalf("error scanning rows")
				}
				err = rows.Close()
				if err != nil {
					b.Fatalf("error closing rows")
				}
			}
		})

		b.Run("multiple-rows", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var users []User
				rows, err := sqlDB.Queryx(`SELECT id, name, age FROM users OFFSET $1 LIMIT 10`, i%90)
				if err != nil {
					b.Fatalf("query error: %s", err.Error())
				}
				for j := 0; j < 10; j++ {
					if !rows.Next() {
						b.Fatalf("missing user from inserted record, offset: %d", i%90)
					}
					var user User
					err = rows.Scan(&user.ID, &user.Name, &user.Age)
					if err != nil {
						b.Fatalf("error scanning rows")
					}
					users = append(users, user)
				}
				if len(users) < 10 {
					b.Fatalf("expected 10 scanned users, but got: %d", len(users))
				}

				err = rows.Close()
				if err != nil {
					b.Fatalf("error closing rows")
				}
			}
		})
	})

	b.Run("sqlx", func(b *testing.B) {
		sqlxDB, err := sqlx.Open(driver, connStr)
		if err != nil {
			b.Fatalf("error creating sqlx client: %s", err)
		}
		sqlxDB.SetMaxOpenConns(1)

		err = recreateTable(connStr)
		if err != nil {
			b.Fatalf("error creating table: %s", err.Error())
		}

		err = insertUsers(connStr, 100)
		if err != nil {
			b.Fatalf("error inserting users: %s", err.Error())
		}

		b.Run("single-row", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var user User
				rows, err := sqlxDB.Queryx(`SELECT * FROM users OFFSET $1 LIMIT 1`, i%100)
				if err != nil {
					b.Fatalf("query error: %s", err.Error())
				}
				if !rows.Next() {
					b.Fatalf("missing user from inserted record, offset: %d", i%100)
				}
				err = rows.StructScan(&user)
				if err != nil {
					b.Fatalf("error scanning rows")
				}
				err = rows.Close()
				if err != nil {
					b.Fatalf("error closing rows")
				}
			}
		})

		b.Run("multiple-rows", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var users []User
				rows, err := sqlxDB.Queryx(`SELECT * FROM users OFFSET $1 LIMIT 10`, i%90)
				if err != nil {
					b.Fatalf("query error: %s", err.Error())
				}
				for j := 0; j < 10; j++ {
					if !rows.Next() {
						b.Fatalf("missing user from inserted record, offset: %d", i%90)
					}
					var user User
					rows.StructScan(&user)
					if err != nil {
						b.Fatalf("error scanning rows")
					}
					users = append(users, user)
				}
				if len(users) < 10 {
					b.Fatalf("expected 10 scanned users, but got: %d", len(users))
				}

				err = rows.Close()
				if err != nil {
					b.Fatalf("error closing rows")
				}
			}
		})
	})

	b.Run("gorm", func(b *testing.B) {
		gormDB, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
		if err != nil {
			b.Fatalf("error creating gorm client: %s", err)
		}

		err = recreateTable(connStr)
		if err != nil {
			b.Fatalf("error creating table: %s", err.Error())
		}

		err = insertUsers(connStr, 100)
		if err != nil {
			b.Fatalf("error inserting users: %s", err.Error())
		}

		b.Run("single-row", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var user User
				err := gormDB.Table("users").Offset(i % 100).Take(&user).Error
				if err != nil {
					b.Fatalf("query error: %s", err)
				}
			}
		})

		b.Run("multiple-rows", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var users []User
				err := gormDB.Table("users").Offset(i % 90).Limit(10).Find(&users).Error
				if err != nil {
					b.Fatalf("query error: %s", err.Error())
				}
				if len(users) < 10 {
					b.Fatalf("expected 10 scanned users, but got: %d", len(users))
				}
			}
		})
	})
}

func recreateTable(connStr string) error {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	db.Exec(`DROP TABLE users`)

	_, err = db.Exec(`CREATE TABLE users (
		  id serial PRIMARY KEY,
			age INT,
			name VARCHAR(50)
		)`)
	if err != nil {
		return fmt.Errorf("failed to create new users table: %s", err.Error())
	}

	return nil
}

func insertUsers(connStr string, numUsers int) error {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	for i := 0; i < numUsers; i++ {
		_, err = db.Exec(`INSERT INTO users (name, age) VALUES ($1, $2)`, strconv.Itoa(i), i)
		if err != nil {
			return fmt.Errorf("failed to insert new user: %s", err.Error())
		}
	}

	return nil
}
