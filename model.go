package main

import (
	"database/sql"
	"errors"
)

type company struct {
	ID   int    `json:"id"`
	NAME string `json:"name"`
}

func (c *company) getCompanies(db *sql.DB) error {
	return errors.New("Not implemented")
}

func getCompanies(db *sql.DB, start, count int) ([]company, error) {
	return nil, errors.New("Not implemented")
}
