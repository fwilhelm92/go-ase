// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

// This example shows how to use sql.Nulltypes using the
// database/sql interface and the pure go driver.
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/SAP/go-ase/libase/libdsn"
	_ "github.com/SAP/go-ase/purego"
)

func main() {
	err := DoMain()
	if err != nil {
		log.Printf("Failed: %v", err)
		os.Exit(1)
	}
}

func DoMain() error {
	dsn, err := libdsn.NewInfoFromEnv("")
	if err != nil {
		return fmt.Errorf("error reading DSN info from env: %w", err)
	}

	fmt.Println("Opening database")
	db, err := sql.Open("ase", dsn.AsSimple())
	if err != nil {
		return fmt.Errorf("failed to open connection to database: %w", err)
	}
	defer db.Close()

	if _, err = db.Exec("if object_id('nullTypes') is not null drop table nullTypes"); err != nil {
		return fmt.Errorf("failed to drop table 'nullTypes': %w", err)
	}

	fmt.Println("Creating table 'nullTypes'")
	if _, err = db.Exec("create table nullTypes (a BigInt null)"); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	var samples = []sql.NullInt64{
		{Int64: 123, Valid: true},
		{Valid: false},
		{Int64: 321, Valid: true},
	}

	for _, sample := range samples {
		val, err := sample.Value()
		if err != nil {
			return fmt.Errorf("failed to evaluate sample: %w", err)
		}
		fmt.Printf("Writing a=%v to table\n", val)
		if _, err = db.Exec("insert into nullTypes values (?)", val); err != nil {
			fmt.Println(err)
			return fmt.Errorf("failed to insert values: %w", err)
		}
	}

	fmt.Println("Querying values from table")
	rows, err := db.Query("select * from nullTypes")
	if err != nil {
		return fmt.Errorf("querying failed: %w", err)
	}
	defer rows.Close()

	fmt.Println("Displaying results of query")
	colName, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to retrieve column names: %w", err)
	}

	fmt.Printf("| %-10s |\n", colName[0])
	format := "| %-10v |\n"

	var ni sql.NullInt64

	for rows.Next() {
		if err = rows.Scan(&ni); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		value, _ := ni.Value()
		fmt.Printf(format, value)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error reading rows: %w", err)
	}

	return nil
}
