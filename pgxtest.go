package pgxtest

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var sharedDB *pgx.Conn

func tcheck(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func DB(t testing.TB, migrations []string) *pgxpool.Pool {
	var (
		ctx = context.Background()
		db  = create(ctx, t)
	)
	for _, mig := range migrations {
		_, err := db.Exec(ctx, mig)
		tcheck(t, err)
	}
	return db
}

func Setup(m *testing.M) {
	flag.Parse()
	connect()
	code := m.Run()
	cleanup()
	os.Exit(code)
}

func connect() {
	const url = "postgres:///postgres"
	var err error
	sharedDB, err = pgx.Connect(context.Background(), url)
	check(err)
}

func create(ctx context.Context, t testing.TB) *pgxpool.Pool {
	var (
		n = name()
		q = fmt.Sprintf(`CREATE DATABASE %s`, qi(n))
	)
	_, err := sharedDB.Exec(ctx, q)
	tcheck(t, err)
	p, err := pgxpool.Connect(ctx, fmt.Sprintf("postgres:///%s", n))
	tcheck(t, err)
	return p
}

func name() string {
	var name string
	const b58 = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	for i := 0; i < 16; i++ {
		name += string(b58[rand.Intn(len(b58))])
	}
	return "pgxtest_" + name
}

func cleanup() {
	ctx := context.Background()
	const q1 = `
		select datname
		from pg_database
		where datname like 'pgxtest_%'
	`
	var dbs []string
	rows, err := sharedDB.Query(ctx, q1)
	check(err)

	defer rows.Close()
	for rows.Next() {
		var d string
		check(rows.Scan(&d))
		dbs = append(dbs, d)
	}
	check(rows.Err())

	for _, name := range dbs {
		var q2 = fmt.Sprintf(`DROP DATABASE IF EXISTS %s WITH (FORCE)`, qi(name))
		_, err = sharedDB.Exec(ctx, q2)
		check(err)
	}
}

func qi(s string) string {
	return pgx.Identifier{s}.Sanitize()
}
