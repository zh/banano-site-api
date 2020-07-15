package main

import (
	"crypto/subtle"
	"encoding/json"
	"log"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/spf13/viper"
	"go.etcd.io/bbolt"
)

// App Application hold
type App struct {
	Router   *mux.Router
	Config   *viper.Viper
	Database *bbolt.DB
}

// Initialize connect to the database
func (a *App) Initialize(db *bbolt.DB, cfg *viper.Viper) {
	a.Router = mux.NewRouter()
	a.Config = cfg
	a.Database = db
	a.initializeRoutes()
}

// Middleware
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func BasicAuth(handler http.HandlerFunc, username, password, realm string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()

		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
			respondWithError(w, http.StatusUnauthorized, "Not Authorized")
			return
		}

		handler(w, r)
	}
}

func (a *App) getAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	_, err := regexp.MatchString(`0-9A-Za-z_`, vars["username"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid account")
		return
	}

	account, err := LoadAccount(a.Database, []byte(vars["username"]))
	if err == errAccountNotFound {
		respondWithError(w, http.StatusNotFound, "Account not found")
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, account)
}

func (a *App) getAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := LoadAccounts(a.Database)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, accounts)
}

func (a *App) createAccount(w http.ResponseWriter, r *http.Request) {
	var p Account
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	err := p.CreateAccount(a.Database)
	if err == errAccountNotUniq {
		respondWithError(w, http.StatusNotFound, "Account already exists")
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, p)
}

func (a *App) updateAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	_, err := regexp.MatchString(`0-9A-Za-z_`, vars["username"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid account")
		return
	}

	// get payload from params
	var p Account
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if p.Address == "" {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	// prepare upload payload
	pp := Account{Username: vars["username"], Address: p.Address}
	if err := pp.UpdateAccount(a.Database); err != nil {
		switch err {
		case errAccountNotFound:
			respondWithError(w, http.StatusNotFound, "Account not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, pp)
}

func (a *App) deleteAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	_, err := regexp.MatchString(`0-9A-Za-z_`, vars["username"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid account")
		return
	}

	p := Account{Username: vars["username"], Address: ""}
	if err := p.DeleteAccount(a.Database); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (a *App) initializeRoutes() {
	user := a.Config.Get("AppUser").(string)
	pass := a.Config.Get("AppPass").(string)
	msg := "Provide user and pass"
	a.Router.HandleFunc("/u/{username:[0-9A-Za-z_]+}", a.getAccount).Methods("GET")
	a.Router.HandleFunc("/api/v1/account", a.createAccount).Methods("POST")
	// need auth
	a.Router.HandleFunc("/api/v1/accounts", BasicAuth(a.getAccounts, user, pass, msg)).Methods("GET")
	a.Router.HandleFunc("/api/v1/account/{username:[0-9A-Za-z_]+}", BasicAuth(a.updateAccount, user, pass, msg)).Methods("PUT")
	a.Router.HandleFunc("/api/v1/account/{username:[0-9A-Za-z_]+}", BasicAuth(a.deleteAccount, user, pass, msg)).Methods("DELETE")
}

// Run start the application
func (a *App) Run(addr string) {
	handler := cors.Default().Handler(limit(a.Router))
	log.Fatal(http.ListenAndServe(addr, handler))
}
