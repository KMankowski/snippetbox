package main

import (
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/KMankowski/snippetbox/internal/models"
	_ "github.com/go-sql-driver/mysql"
	"github.com/justinas/alice"
)

type application struct {
	logger        *slog.Logger
	db            models.SnippetModel
	templateCache map[string]*template.Template
}

func main() {
	// Initialize flags
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-") {
			fmt.Printf("Flag: %s\n", arg)
		} else {
			fmt.Printf("Argument: %s\n", arg)
		}
	}
	fmt.Print("\n")
	addr := flag.String("addr", ":4040", "Port for the server to listen on")
	sqlDataSourceName := flag.String("dsn", "", "REQUIRED data source name for mysql driver")
	flag.Parse()

	// Validate flags
	if sqlDataSourceName == nil || *sqlDataSourceName == "" {
		fmt.Println("REQUIRED dsn flag argument not provided")
		os.Exit(2)
	}

	// Initialize logger
	// A single logger is thread-safe! Internally uses mutex or smth. Very good for net/http which is threaded by default
	// os.Stdout makes sense because then the environment can configure treatment of logging w/out being coupled to source code
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))

	// Initialize DB
	dbConnectionPool, err := openDB(*sqlDataSourceName)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	// `defer` statements are not run when os.Exit() is hit or ctrl+c is used to interrupt the process
	// So this is a bit superfluous, but still a good idea, especially if graceful shutdown is added in later changes
	defer dbConnectionPool.Close()

	// Parsed template/HTML page cache
	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(2)
	}

	// Initialize application configuration
	app := application{
		logger:        logger,
		db:            models.SnippetModel{DBConnectionPool: dbConnectionPool},
		templateCache: templateCache,
	}

	// Run the server
	app.logger.Info("starting server", slog.String("addr", *addr))
	// err = http.ListenAndServe(*addr, app.handlePanic(app.logRequest(setCommonHeaders(app.getRouting()))))
	defaultMiddleware := alice.New(app.handlePanic, app.logRequest, setCommonHeaders)
	err = http.ListenAndServe(*addr, defaultMiddleware.Then(app.getRouting()))
	app.logger.Error(err.Error())
	os.Exit(3)
}

func (app *application) getRouting() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", app.home)
	mux.HandleFunc("GET /snippet/view/{id}", app.snippetView)
	mux.HandleFunc("GET /snippet/create", app.snippetCreate)
	mux.HandleFunc("POST /snippet/create", app.snippetCreatePost)

	fs := http.FileServer(http.Dir("./ui/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	return mux
}

func openDB(dsn string) (*sql.DB, error) {
	// Open() does not actually create any connections to the database, that is done lazily (as needed)
	// This just initializes the pool
	// To check that it was initialized without issue, we use Ping() to create a connection and check for any errors
	dbConnectionPool, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	err = dbConnectionPool.Ping()
	if err != nil {
		dbConnectionPool.Close()
		return nil, err
	}

	return dbConnectionPool, err
}
