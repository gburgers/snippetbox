package main

import (
	"database/sql"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/gburgers/snippetbox/internal/models"

	// "github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/postgresstore" // instead of mysqlstore
	// instead of mysqlstore

	"github.com/alexedwards/scs/v2"

	// _ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// Define an application struct to hold the application-wide dependencies for the
// web application. For now we'll only include the structured logger, but we'll
// add more to this as the build progresses.
type application struct {
	logger         *slog.Logger
	snippets       *models.SnippetModel
	templateCache  map[string]*template.Template
	sessionManager *scs.SessionManager
}

func main() {
	// Define a new command-line flag with the name 'addr', a default value of ":4000"
	// and some short help text explaining what the flag controls. The value of the
	// flag will be stored in the addr variable at runtime.
	addr := flag.String("addr", ":4000", "HTTP network address")

	// Importantly, we use the flag.Parse() function to parse the command-line flag.
	// This reads in the command-line flag value and assigns it to the addr
	// variable. You need to call this *before* you use the addr variable
	// otherwise it will always contain the default value of ":4000". If any errors are
	// encountered during parsing the application will be terminated.

	// mysql:
	// dsn := flag.String("dsn", "root:pass@(sb-db-svc:3306)/snippetbox?parseTime=true", "MySQL data source name")

	// Check LOCAL OR K3S POSTGRESQL DSN!!!!!!

	// !!!!postgresQL localhost docker-compose!!!!
	dsn := flag.String("dsn", "postgresql://sb-user:pass@localhost:5432/sb-database?sslmode=disable", "PostgreSQL data source name")

	// kubernetes k3s
	// dsn := flag.String("dsn", os.Getenv("DATABASE_DSN"), "PostgreSQL data source name")

	flag.Parse()
	// Use the slog.New() function to initialize a new structured logger, which
	// writes to the standard out stream and uses the default settings.
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))

	// The sql.Open() function initializes a new sql.DB object, which is essentially a
	// pool of database connections.
	// To keep the main() function tidy I've put the code for creating a connection
	// pool into the separate openDB() function below. We pass openDB() the DSN
	// from the command-line flag.
	db, err := openDB(*dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// We also defer a call to db.Close(), so that the connection pool is closed
	// before the main() function exits.
	defer db.Close()

	// Initialize a new template cache...
	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Use the scs.New() function to initialize a new session manager. Then we
	// configure it to use our MySQL database as the session store, and set a
	// lifetime of 12 hours (so that sessions automatically expire 12 hours
	// after first being created).
	sessionManager := scs.New()
	// sessionManager.Store = mysqlstore.New(db)
	sessionManager.Store = postgresstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour

	// Initialize a new instance of our application struct, containing the
	// dependencies (for now, just the structured logger).
	app := &application{
		logger:         logger,
		snippets:       &models.SnippetModel{DB: db},
		templateCache:  templateCache,
		sessionManager: sessionManager,
	}

	// Use the Info() method to log the starting server message at Info severity
	// (along with the listen address as an attribute).
	logger.Info("starting server", slog.String("addr", *addr))

	// Call the new app.routes() method to get the servemux containing our routes,
	// and pass that to http.ListenAndServe().
	// err := http.ListenAndServe(*addr, mux)
	err = http.ListenAndServe(*addr, app.routes())
	// And we also use the Error() method to log any error message returned by
	// http.ListenAndServe() at Error severity (with no additional attributes),
	// and then call os.Exit(1) to terminate the application with exit code 1.
	logger.Error(err.Error())
	os.Exit(1)
}

// The openDB() function wraps sql.Open() and returns a sql.DB connection pool
// for a given DSN.
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
