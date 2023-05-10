package api 

import (
    "os"
    "log"
    "encoding/json"
    "net/http"
    "fmt"
    "strconv"

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

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) error {
    if r.Method != "POST" {
        return WriteJSON(w, http.StatusBadRequest, "this method is not supported you should use POST instead")
    }

    req := new(types.LoginRequest)
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        return err
    }

    acc, err := s.store.GetAccountByNumber(int64(req.Number))
    if err != nil {
        return err
    }

    if !acc.ValidatePassword(req.Password) {
        return WriteJSON(w, http.StatusForbidden, "Either number or password is incorect")
    }

    token, err := createJWT(acc)
    if err != nil {
        return err
    }

    resp := types.LoginResponse{
        Token: token,
        Number: acc.Number,
    }

    return WriteJSON(w, http.StatusOK, resp)
}

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
    if r.Method == "GET" {
        return s.handleGetAccount(w, r)
    }
    
    if r.Method == "POST" {
        return s.handleCreateAccount(w, r)
    }

    
    return fmt.Errorf("method not allowed %s", r.Method)
}
   
func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
    accounts, err := s.store.GetAccounts()
    if err != nil {
        return err
    }

    return WriteJSON(w, http.StatusOK, accounts)
}

func (s *APIServer) handleAccountWithID(w http.ResponseWriter, r *http.Request) error {
    if r.Method == "GET" {
        return s.handleGetAccountByID(w, r)
    }
    if r.Method == "DELETE" {
        return s.handleDeleteAccount(w, r)
    }

    return fmt.Errorf("method not allowed %s", r.Method)
}


func (s *APIServer) handleGetAccountByID(w http.ResponseWriter, r *http.Request) error {
        id, err := getID(r)

        if err != nil {
            return err
        }

        account, err := s.store.GetAccountByID(id)

        if err != nil {
            return err
        }

        return WriteJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
    createAccountReq := new(types.CreateAccountRequest)
    if err := json.NewDecoder(r.Body).Decode(createAccountReq); err != nil {
        return err
    }

    account, err := types.NewAccount(createAccountReq.FirstName, createAccountReq.LastName, createAccountReq.Password)

    if err != nil {
        return err 
    }

    if err := s.store.CreateAccount(account); err != nil {
        return err
    }


    return WriteJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
    id, err := getID(r)

    if err != nil {
        return err
    }
    
    if err := s.store.DeleteAccount(id); err != nil {
        return err
    }

    return WriteJSON(w, http.StatusOK, map[string]int{"deleted": id})
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
       

        userID, err := getID(r)
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


func validateJWT(token string) (*jwt.Token, error) {
    secret := os.Getenv("JWT_SECRET")
    return jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
        }

        return []byte(secret), nil
    })
}

func createJWT(account *types.Account) (string, error) {
    claims := &jwt.MapClaims{
        "expiresAt": 15000, 
        "accountNumber": account.Number,
    }

    secret := os.Getenv("JWT_SECRET")
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    return token.SignedString([]byte(secret))
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

func getID(r *http.Request) (int, error) {
	idStr := r.URL.Path[len("/account/"):]
    id, err := strconv.Atoi(idStr)
    if err != nil {
        return id, fmt.Errorf("This id is not a valid integer")
    }
    return id, nil
} 
