// Package support holds what the tests under test/ share: the database and Redis containers of a test
// package, the envs that run the deployables of app, SQL fixtures, and HTTP requests. It holds no behavior of
// the application: the tests run the app's own wiring, through forgo's forgotest.
//
// A test package starts its database and builds its env once, in TestMain, and shares them between its
// tests, which therefore do not call t.Parallel and reset the database with Fixture first:
//
//	var env *forgotest.Env
//
//	func TestMain(m *testing.M) {
//		stop := support.StartDatabase()
//		defer stop()
//		env = support.HTTP()
//		defer env.Close()
//		m.Run()
//	}
package support

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/efegreat/forgo-adapter/rabbitmq"
	"github.com/efegreat/forgo-adapter/redis/locker"
	"github.com/efegreat/forgo/cli"
	"github.com/efegreat/forgo/forgotest"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"

	"example/app"
	"example/config"
	"example/test/fixtures"
)

// db is the database of the test package, for fixtures and checks.
var db *sql.DB

// StartDatabase starts the PostgreSQL container of the test package, points DATABASE_URL at it, and
// migrates it with the migrate command of app, as a deploy would. It returns the function that
// stops it. Call it first in TestMain, since config.DB reads DATABASE_URL when an env is built. A failure
// ends the test binary, and testcontainers' reaper removes the container.
func StartDatabase() (stop func()) {
	ctx := context.Background()
	ctr, err := tcpostgres.Run(ctx, "postgres:18",
		tcpostgres.WithDatabase("example"),
		tcpostgres.WithUsername("example"),
		tcpostgres.WithPassword("example"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		log.Fatalf("support: start postgres: %v", err)
	}
	dsn, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("support: postgres address: %v", err)
	}
	if err := os.Setenv("DATABASE_URL", dsn); err != nil {
		log.Fatal(err)
	}
	if err := app.CLI(cli.WithArgs([]string{"cli", "migrate", "-a=up"})).Run(); err != nil {
		log.Fatalf("support: migrate: %v", err)
	}
	if db, err = sql.Open("postgres", dsn); err != nil {
		log.Fatal(err)
	}
	return func() {
		_ = db.Close()
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			log.Printf("support: stop postgres: %v", err)
		}
	}
}

// redisAddr is the address of the test package's Redis, for HoldLock.
var redisAddr string

// StartRedis starts the Redis container of the test package, which holds the locks of the scheduled jobs,
// and points REDIS_ADDR at it. It returns the function that stops it. Call it in TestMain before an env of
// the cron deployable is built, since config.Redis reads REDIS_ADDR then.
func StartRedis() (stop func()) {
	ctx := context.Background()
	ctr, err := tcredis.Run(ctx, "redis:8")
	if err != nil {
		log.Fatalf("support: start redis: %v", err)
	}
	if redisAddr, err = ctr.Endpoint(ctx, ""); err != nil {
		log.Fatalf("support: redis address: %v", err)
	}
	if err := os.Setenv("REDIS_ADDR", redisAddr); err != nil {
		log.Fatal(err)
	}
	return func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			log.Printf("support: stop redis: %v", err)
		}
	}
}

// HoldLock takes the lock key for ttl in the test package's Redis, as another replica that runs the job
// would, and releases it when t ends.
func HoldLock(t testing.TB, key string, ttl time.Duration) {
	t.Helper()
	cl := redis.NewClient(&redis.Options{Addr: redisAddr})
	other := locker.NewRedis(cl)
	if ok, err := other.Lock(t.Context(), key, ttl); err != nil || !ok {
		t.Fatalf("lock %s: %t, %v", key, ok, err)
	}
	t.Cleanup(func() {
		_ = other.Unlock(context.Background(), key)
		_ = cl.Close()
	})
}

// DB returns the database of the test package, for tests of repositories.
func DB() *sql.DB {
	return db
}

// HTTP builds the env of the orders API, as cmd/http runs it, and ends the test binary when it fails.
func HTTP() *forgotest.Env {
	return mustEnv(forgotest.New(app.HTTP()))
}

// GRPC builds the env of the orders API over gRPC, as cmd/grpc runs it, and ends the test binary when it
// fails.
func GRPC() *forgotest.Env {
	return mustEnv(forgotest.New(app.GRPC()))
}

// Worker builds the env of the worker, as cmd/worker runs it, with forgotest's stand-in in place of
// RabbitMQ: tests run no broker.
func Worker() *forgotest.Env {
	return mustEnv(forgotest.New(app.Worker(),
		forgotest.ReplaceResource[rabbitmq.Config](forgotest.Broker(config.Broker().Name))))
}

// Cron builds the env of the scheduled jobs, as cmd/cron runs it. It needs StartRedis, for the locks of the
// jobs.
func Cron() *forgotest.Env {
	return mustEnv(forgotest.New(app.Cron()))
}

// Command runs the command of the CLI that args name, such as "retry-failed", "-subscriber=order", as
// cmd/cli runs it.
func Command(args ...string) error {
	return app.CLI(cli.WithArgs(append([]string{"cli"}, args...))).Run()
}

func mustEnv(env *forgotest.Env, err error) *forgotest.Env {
	if err != nil {
		log.Fatalf("support: %v", err)
	}
	return env
}

// migrationsTable is the table of the applied migrations, which Fixture keeps.
const migrationsTable = "migrations"

// Fixture empties every table but the migrations, then runs the fixtures names of test/fixtures in order,
// such as "orders/placed_order.sql". Call it first in each test, with no names for an empty database: the
// tests of a package share the database, and the outbox and the inbox are tables too.
func Fixture(t testing.TB, names ...string) {
	t.Helper()
	ctx := t.Context()
	rows, err := db.QueryContext(ctx,
		`SELECT tablename FROM pg_tables WHERE schemaname = current_schema() AND tablename <> $1`, migrationsTable)
	if err != nil {
		t.Fatal(err)
	}
	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			t.Fatal(err)
		}
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(tables) > 0 {
		if _, err := db.ExecContext(ctx, "TRUNCATE "+strings.Join(tables, ", ")); err != nil {
			t.Fatal(err)
		}
	}
	Load(t, names...)
}

// Load runs the fixtures names of test/fixtures in order, over what the database holds, such as an order
// that a test creates after its message failed.
func Load(t testing.TB, names ...string) {
	t.Helper()
	ctx := t.Context()
	for _, name := range names {
		query, err := fixtures.FS.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.ExecContext(ctx, string(query)); err != nil {
			t.Fatalf("fixture %s: %v", name, err)
		}
	}
}

// Count returns the number that query, a SELECT COUNT(*) with args, counts.
func Count(t testing.TB, query string, args ...any) int {
	t.Helper()
	var n int
	if err := db.QueryRowContext(t.Context(), query, args...).Scan(&n); err != nil {
		t.Fatal(fmt.Errorf("%s: %w", query, err))
	}
	return n
}
