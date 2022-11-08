//Filename: cmd/api/main.go

package main

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"time"

	_ "github.com/lib/pq"
	"grocery.jamesfaber.net/internal/data"
	"grocery.jamesfaber.net/internal/jsonlog"
)

// The application version nimber
const version = "1.0.0"

// The configuration settings
// The config struct -a set of complex port properties that specify the data type of the complex data type elements or the schema of the data
type config struct {
	port int
	env  string //development, staging, production, etc
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	limiter struct {
		rps     float64 // requests/second
		burst   int
		enabled bool
	}
}

// Dependency injection - the process of supplying a resource that a given piece of code requires.
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
}

func main() {
	var cfg config
	// read in the flags that are needed to populate our config
	// a flag is a predefined bit or bit sequence that holds a binary value.(not sure)
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development | staging | production )")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GROCERY_DB_DSN"), "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	// These are flags for the rate limiter
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	// To parse -is where a string of commands – usually a program – is separated into more easily processed components, which are analyzed for correct syntax and then attached to tags that define each component.
	flag.Parse()

	//Create a logger - Logging is a means of tracking events that happen when some software runs.
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	// Create the connection pool
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	defer db.Close()
	// Log the successful connection pool
	logger.PrintInfo("database connection pool established", nil)

	//Create an instance of our applications struct
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}
	// cal app.Serve() to start the serve
	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
}

// The openDB() function returns a *sql.DB connection pool
func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)
	// Create a context with a 5-second timeout deadline
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil

}
