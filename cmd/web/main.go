package main

import (
   "crypto/tls"
   "flag"
   "log"
   "net/http"
   "os"
   "database/sql"
   "html/template"
   "time"

   "cb.net/snippetbox/pkg/models/mysql"
   "cb.net/snippetbox/pkg/models"
   _"github.com/go-sql-driver/mysql"
   "github.com/golangcollege/sessions"

)

type contextKey string

var contextKeyIsAuthenticated = contextKey("isAuthenticated")

type application struct {
   errorLog *log.Logger
   infoLog *log.Logger
   session *sessions.Session
   snippets interface {
      Insert(string, string, string) (int, error)
      Get(int) (*models.Snippet, error)
      Latest() ([]*models.Snippet, error)
   }
   templateCache map[string]*template.Template
   users interface {
      Insert(string, string, string) error
      Authenticate(string, string) (int, error)
      Get(int) (*models.User, error)
      ChangePassword(int, string, string) error
   }
}

//Config struct for flags
type Config struct {
   Addr string
   StaticDir string
}

func main() {
   //using a struct for storing variables
   cfg := new(Config)
   flag.StringVar(&cfg.Addr, "addr", ":4000", "HTTP network address")
   flag.StringVar(&cfg.StaticDir, "static-dir", "./ui/static", "Path to static assets")
   
   // Define a new command-line flag with the name 'addr', a default value of ":4000"
   // and some short help text explaining what the flag controls. The value of the
   // flag will be stored in the addr variable at runtime.
   //addr := flag.String("addr", ":4000", "HTTP network address")
   dsn := flag.String("dsn", "web:P@ssw0rd@/snippetbox?parseTime=true", "MySQL data source name")

   // Define a new command-line flag for the session secret (a random key which
   // will be used to encrypt and authenticate session cookies). It should be 32
   // will be used to encrypt and authenticate session cookies). It should be 32
   // bytes long.
   secret := flag.String("secret", "s6Ndh+pPbnzHbS*+9Pk8qGWhTzbpa@ge", "Secret key")
   
   // Importantly, we use the flag.Parse() function to parse the command-line flag.
   // This reads in the command-line flag value and assigns it to the addr
   // variable. You need to call this *before* you use the addr variable
   // otherwise it will always contain the default value of ":4000". If any errors are
   // encountered during parsing the application will be terminated.
   flag.Parse()

   // Use log.New() to create a logger for writing information messages. This takes
   // three parameters: the destination to write the logs to (os.Stdout), a string
   // prefix for message (INFO followed by a tab), and flags to indicate what
   // additional information to include (local date and time). Note that the flags
   // are joined using the bitwise OR operator |.
   infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

   // Create a logger for writing error messages in the same way, but use stderr as
   // the destination and use the log.Lshortfile flag to include the relevant
   // file name and line number.
   errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

   db, err := openDB(*dsn)
   if err != nil {
      errorLog.Fatal(err)
   }

   defer db.Close()

   //initialize a new template cache...
   templateCache, err := newTemplateCache("./ui/html/")
   if err != nil {
      errorLog.Fatal(err)
   }

   // Use the sessions.New() function to initialize a new session manager,
   // passing in the secret key as the parameter. Then we configure it so
   // sessions always expires after 12 hours.
   session := sessions.New([]byte(*secret))
   session.Lifetime = 12 * time.Hour
   session.Secure = true

   //application dependencies
   app := &application{
       errorLog: errorLog,
       infoLog: infoLog,
       session: session,
       snippets: &mysql.SnippetModel{DB: db},
       templateCache: templateCache,
       users: &mysql.UserModel{DB: db},
   }

   // Initialize a tls.Config struct to hold the non-default TLS settings we want
   // the server to use.
   tlsConfig := &tls.Config{
      PreferServerCipherSuites: true,
      CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
      MinVersion: tls.VersionTLS12,
      MaxVersion: tls.VersionTLS12,

   }   

   //initialize servemux
   mux := http.NewServeMux()

   // Create a file server which serves files out of the "./ui/static" directory.
   // Note that the path given to the http.Dir function is relative to the project
   // directory root.
   fileServer := http.FileServer(http.Dir("./ui/static/"))

   // Use the mux.Handle() function to register the file server as the handler for
   // all URL paths that start with "/static/". For matching paths, we strip the
   // "/static" prefix before the request reaches the file server.
   mux.Handle("/static/", http.StripPrefix("/static", fileServer))

   // Initialize a new http.Server struct. We set the Addr and Handler fields so
   // that the server uses the same network address and routes as before, and set
   // the ErrorLog field so that the server now uses the custom errorLog logger in
   // the event of any problems.
   srv := &http.Server{
      Addr: cfg.Addr,
      ErrorLog: errorLog,
      Handler: app.routes(),
      TLSConfig: tlsConfig,
      // Add Idle, Read and Write timeouts to the server.
      IdleTimeout: time.Minute,
      ReadTimeout: 5 * time.Second,
      WriteTimeout: 10 * time.Second,
   }

   infoLog.Printf("Starting server on %s", cfg.Addr)

   // Use the ListenAndServeTLS() method to start the HTTPS server. We
   // pass in the paths to the TLS certificate and corresponding private key as
   // the two parameters.
   //err = srv.ListenAndServe()
   err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
   errorLog.Fatal(err)
}

func openDB(dsn string) (*sql.DB, error) {
   db, err := sql.Open("mysql", dsn)
   if err != nil {
      return nil, err
   }
   if err = db.Ping(); err != nil {
      return nil, err
   }
   return db, nil
}
