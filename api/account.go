package api

import (
    "net/http"
    "fmt"
    "gobank/types"
    "encoding/json"
    "strconv"
)



func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
    if r.Method == "GET" {
        return s.handleGetAccount(w, r)
    }

    if r.Method == "POST" {
        return s.handleCreateAccount(w, r)
    }


    return fmt.Errorf("method not allowed %s", r.Method)
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

func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
    accounts, err := s.store.GetAccounts()
    if err != nil {
        return err
    }

    return WriteJSON(w, http.StatusOK, accounts)
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



func getID(r *http.Request) (int, error) {
	idStr := r.URL.Path[len("/account/"):]
    id, err := strconv.Atoi(idStr)
    if err != nil {
        return id, fmt.Errorf("This id is not a valid integer")
    }
    return id, nil
}

