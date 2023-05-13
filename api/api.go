package api

import (
    "log"
    "encoding/json"
    "net/http"
    "fmt"
    "time"
    jwt "github.com/golang-jwt/jwt/v4"
    "gobank/storage"
    "gobank/types"
)

type APIServer struct {
    listenAddr string
    store storage.Storage
}

func NewApiServer(listenAddr string, store storage.Storage) *APIServer {
    return &APIServer {
        listenAddr: listenAddr,
        store: store,
    }
}

func (s *APIServer) Run() error {
    router := http.NewServeMux()

    router.HandleFunc("/login", makeHTTPHandleFunc(s.handleLogin))
    router.HandleFunc("/account", makeHTTPHandleFunc(s.handleAccount))
    router.HandleFunc("/account/", withJWTAuth(makeHTTPHandleFunc(s.handleAccountWithID), s.store))
    router.HandleFunc("/transfer", makeHTTPHandleFunc(s.handleTransfer))

    log.Println("json API server running on port: ", s.listenAddr)

    server := &http.Server{
		Addr:         s.listenAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	return server.ListenAndServe()
}



func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
    if r.Method == "POST" {
        transferReq := new(types.TransferRequest)
        if err := json.NewDecoder(r.Body).Decode(transferReq); err != nil {
            return err
        }
        defer r.Body.Close()

        return WriteJSON(w, http.StatusOK, transferReq)
    }
    return fmt.Errorf("method %s  not supported, you should use POST instead", r.Method)
}


func WriteJSON(w http.ResponseWriter, status int, v any) error {
    w.Header().Add("Content-Type", "application/json")
    w.WriteHeader(status)
    return json.NewEncoder(w).Encode(v)
}

func withJWTAuth(handlerFunc http.HandlerFunc, s storage.Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        log.Println("calling JWT auth middleware")

        userID, err := getID(r)
        if err != nil {
            WriteJSON(w, http.StatusForbidden, ApiError{Error: "permission denied"})
            return
        }

        tokenString := r.Header.Get("x-jwt-token")
        token, err := validateJWT(tokenString)
        if err != nil {
            WriteJSON(w, http.StatusForbidden, ApiError{Error: "permission denied"})
            return
        }
        if !token.Valid {
            WriteJSON(w, http.StatusForbidden, ApiError{Error: "permission denied"})
            return
        }

        if err != nil {
            WriteJSON(w, http.StatusForbidden, ApiError{Error: "permission denied"})
            return
        }

        account, err := s.GetAccountByID(userID)
        if err != nil {
            WriteJSON(w, http.StatusBadRequest, ApiError{Error: "This account does not exist"})
            return
        }

        claims := token.Claims.(jwt.MapClaims)
        if account.Number != int64(claims["accountNumber"].(float64)) {
            WriteJSON(w, http.StatusForbidden, ApiError{Error: "permission denied"})
            return
        }

        fmt.Println(claims["accountNumber"])

        handlerFunc(w, r)
    }
}



type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
    Error string `json:"error"`
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if err := f(w, r); err != nil {
            // handle the error
            WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
        }
    }
}



