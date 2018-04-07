package main

import (
	"database/sql"
	"errors"
)

type company struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func (c *company) getCompany(db *sql.DB) error {
	return db.QueryRow("SELECT id, name FROM company WHERE deleted = false AND id = $1",
		c.Id).Scan(&c.Id, &c.Name)
}

func (c *company) getCompanies(db *sql.DB) error {
	return errors.New("Not implemented")
}

func getCompanies(db *sql.DB, start, count int) ([]company, error) {
	rows, err := db.Query(
		"SELECT id, name FROM company WHERE deleted = false LIMIT $1 OFFSET $2",
		count,
		start)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	companies := []company{}

	for rows.Next() {
		var c company
		if err := rows.Scan(&c.Id, &c.Name); err != nil {
			return nil, err
		}
		companies = append(companies, c)
	}
	return companies, nil
}
