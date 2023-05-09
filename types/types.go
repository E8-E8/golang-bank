package types

import (
	"math/rand"
	"time"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
    Number int64  `json:"number"`
    Password string  `json:"password"`
}

type LoginResponse struct {
    Number int64 `json:"number"`
    Token string `json:"token"`
}

type TransferRequest struct {
    ToAccount int64 `json:"toAccount"` 
    Amount int64 `json:"amount"` 
}

type CreateAccountRequest struct {
    FirstName string `json:"firstName"`
    LastName string `json:"lastName"`
    Password string `json:"password"`
}

type Account struct {
    ID int `json:"id"`
    FirstName string `json:"fistName"`
    LastName string `json:"lastName"`
    EncryptedPassword string `json:"-"`
    Number int64 `json:"number"`
    Balance int64 `json:"balance"`
    CreatedAt time.Time  `json:"createdAt"`
}

func (acc *Account) ValidatePassword(pw string) bool {
    return bcrypt.CompareHashAndPassword([]byte(acc.EncryptedPassword), []byte(pw)) == nil
}

func  NewAccount(firstName, lastName, password string) (*Account, error)  {
    encpw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) 

    if err != nil {
        return nil, err
    }

    return &Account {
        FirstName: firstName,
        LastName: lastName,
        Number: int64(rand.Intn(10000000)),
        EncryptedPassword: string(encpw),
        CreatedAt: time.Now().UTC(),
    }, nil
}


