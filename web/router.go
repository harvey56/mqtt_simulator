package web

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"mqtt-mochi-server/middleware"
	"mqtt-mochi-server/ws"
)

type AppRouter struct {
	Router      *mux.Router
	DB          *sql.DB
	WSHub       *ws.Hub
	RestartChan chan struct{}
}

func NewRouter() *AppRouter {
	ar := &AppRouter{}
	ar.Initialize()
	log.Printf("New AppRouter Initialized")

	return ar
}

func (ar *AppRouter) Initialize() {
	ar.Router = mux.NewRouter()
	ar.Router.NotFoundHandler = NotFoundHandler()
	ar.WSHub = ws.NewHub()
	go ar.WSHub.Run()

	apiV1Router := ar.Router.PathPrefix("/api/v1").Subrouter()
	ar.SetupAPIV1Router("/api/v1", apiV1Router)
}

func (ar *AppRouter) SetDB(db *sql.DB) {
	ar.DB = db
	ar.RestartChan = make(chan struct{}, 1)
	log.Println("Set the DB connection in the router")

	middlewareAppRouter := &middleware.AppRouter{
		Router:      ar.Router,
		DB:          ar.DB,
		RestartChan: ar.RestartChan,
	}
	ar.Router.Use(middleware.AppRouterInjector(middlewareAppRouter))
}

func (ar *AppRouter) SetupAPIV1Router(prefix string, s *mux.Router) {
	ar.Get(s, "/", middleware.GetIndex)
	ar.Post(s, "/messages", middleware.PostMessage)
	ar.Get(s, "/messages", middleware.GetMessages)
	ar.Delete(s, "/messages/{id}", middleware.DeleteMessage)
	ar.Put(s, "/messages/{id}", middleware.PutMessage)
	ar.Get(s, "/messages/{id}", middleware.GetMessageByID)

	ar.Router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(ar.WSHub, w, r)
	})
}

func (ar *AppRouter) Get(s *mux.Router, path string, f func(w http.ResponseWriter, r *http.Request)) {
	s.HandleFunc(path, f).Methods("GET")
}

func (ar *AppRouter) Post(s *mux.Router, path string, f func(w http.ResponseWriter, r *http.Request)) {
	s.HandleFunc(path, f).Methods("POST")
}

func (ar *AppRouter) Delete(s *mux.Router, path string, f func(w http.ResponseWriter, r *http.Request)) {
	s.HandleFunc(path, f).Methods("DELETE")
}

func (ar *AppRouter) Put(s *mux.Router, path string, f func(w http.ResponseWriter, r *http.Request)) {
	s.HandleFunc(path, f).Methods("PUT")
}

func NotFoundHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "404 - Not Found", http.StatusNotFound)
	})
}

func (ar *AppRouter) Run(port string) {
	credentials := handlers.AllowCredentials()
	headers := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type"})
	methods := handlers.AllowedMethods([]string{"POST", "GET", "OPTIONS", "PUT", "DELETE"})
	origins := handlers.AllowedOrigins([]string{"*"})
	handlers.MaxAge(86400)

	go func() {
		err := http.ListenAndServe(port, handlers.CORS(credentials, headers, methods, origins)(ar.Router))
		if err != nil {
			log.Printf("Unable to serve on port %s due to error: %s", port, err)
		}
	}()

	log.Printf("Server listening on port %s", port)
}
