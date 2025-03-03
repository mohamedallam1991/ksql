module github.com/vingarcia/ksql/benchmarks

go 1.16

require (
	github.com/friendsofgo/errors v0.9.2
	github.com/jackc/pgx/v4 v4.13.0
	github.com/jmoiron/sqlx v1.3.4
	github.com/lib/pq v1.10.4
	github.com/vingarcia/ksql v1.4.7
	github.com/vingarcia/ksql/adapters/kpgx v0.0.0-00010101000000-000000000000
	github.com/volatiletech/sqlboiler/v4 v4.12.0
	github.com/volatiletech/strmangle v0.0.4
	gorm.io/driver/postgres v1.2.2
	gorm.io/gorm v1.22.3
)

replace (
	github.com/vingarcia/ksql => ../
	github.com/vingarcia/ksql/adapters/kpgx => ../adapters/kpgx/
)
