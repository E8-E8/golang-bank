package storage

import (
    "database/sql"
    _ "github.com/lib/pq"
)

type Storage interface {
    AccountStorage
}

type PostgresStore struct {
    db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
    connStr := "user=postgres dbname=postgres password=gobank sslmode=disable"
    db, err := sql.Open("postgres", connStr)

    if err != nil {
        return nil, err 
    }

    if err := db.Ping(); err != nil {
        return nil, err
    }

    return &PostgresStore {
        db: db, 
    }, nil
}

func (s *PostgresStore) Init() error {
    return s.CreateAccountTable()
}

func (s *PostgresStore) CreateAccountTable() error {
    query := `create table if not exists account (
        id serial primary key,
        first_name varchar(50),
        last_name varchar(50),
        number serial,
        balance serial,
        encrypted_password varchar(256),
        created_at timestamp
    )`

    _, err := s.db.Exec(query)
    return err
}

