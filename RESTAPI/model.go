package main

import (
	"database/sql"
	"errors"
)

type company struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func (c *company) getCompanies(db *sql.DB) error {
	return errors.New("Not implemented")
}

func getCompanies(db *sql.DB, start, count int) ([]company, error) {
	rows, err := db.Query(
		"SELECT id, name FROM company LIMIT $1 OFFSET $2",
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
