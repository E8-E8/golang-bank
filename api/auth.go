package api

import (
    "net/http"
    "encoding/json"
    "gobank/types"
    "os"
    jwt "github.com/golang-jwt/jwt/v4"
    "fmt"
)


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

