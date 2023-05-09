package storage

import (
    "database/sql"
    "gobank/types"
    "fmt"
)

type AccountStorage interface {
    CreateAccount(*types.Account) error
    DeleteAccount(int) error
    UpdateAccount(*types.Account) error
    GetAccounts() ([]*types.Account, error)
    GetAccountByID(int) (*types.Account, error)
    GetAccountByNumber(int64) (*types.Account, error)
}

func (s *PostgresStore) CreateAccount(acc *types.Account) error  {
    query := `
         insert into account 
         (
             first_name,
             last_name,
             number,
             balance,
             encrypted_password,
             created_at
         )
         values ($1, $2, $3, $4, $5, $6)
    `
    _, err := s.db.Query(
        query,
        acc.FirstName,
        acc.LastName,
        acc.Number,
        acc.Balance,
        acc.EncryptedPassword,
        acc.CreatedAt,
    )
    if err != nil {
        return err
    }

    return nil
}

func (s *PostgresStore) UpdateAccount(*types.Account) error  {
    return nil
}

func (s *PostgresStore) DeleteAccount(id int) error  {
    _, err := s.db.Query(`
        delete from account where id = $1
    `, id)

    return err
}

func (s *PostgresStore) GetAccountByNumber(number int64) (*types.Account, error) {
    rows, err := s.db.Query(`
        select * from account where number = $1
    `, number)

    if err != nil {
        return nil, err
    }

    for rows.Next() {
        return scanIntoAccount(rows)
    }

    return nil, fmt.Errorf("account %d not found", number)
}

func (s *PostgresStore) GetAccountByID(id int ) (*types.Account, error)  {
    rows, err := s.db.Query(`
        select * from account where id = $1
    `, id)
    if err != nil {
        return nil, err
    }
    for rows.Next() {
        return scanIntoAccount(rows)
    }

    return nil, fmt.Errorf("account %d not found", id)
}

func (s *PostgresStore) GetAccounts() ([]*types.Account, error)  {
    rows, err := s.db.Query("select * from account")
    if err != nil {
        return nil, err
    }
    accounts := []*types.Account{}
    for rows.Next() {
        account, err := scanIntoAccount(rows)
        if err != nil {
            return nil, err
        }
        accounts = append(accounts, account)
    }

    return accounts, nil
}

func scanIntoAccount(rows *sql.Rows) (*types.Account, error) {
    account := new(types.Account)
    err := rows.Scan(
        &account.ID,
        &account.FirstName,
        &account.LastName,
        &account.Number,
        &account.Balance,
        &account.EncryptedPassword,
        &account.CreatedAt,
    )
    
    return account, err
}




