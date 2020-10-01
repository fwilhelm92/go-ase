// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/SAP/go-ase/libase/libdsn"
	ase "github.com/SAP/go-ase/purego"
)

// This example shows how to use sql.Nulltypes using the
// database/sql interface, prepared statements and the pure go driver.
func main() {
	if err := DoMain(); err != nil {
		log.Fatalf("nullTypesRaw: %v", err)
	}
}

func DoMain() error {
	dsn, err := libdsn.NewInfoFromEnv("")
	if err != nil {
		return fmt.Errorf("error reading DSN info from env: %w", err)
	}

	db, err := sql.Open("ase", dsn.AsSimple())
	if err != nil {
		return fmt.Errorf("failed to open connection to database: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("error closing db: %v", err)
		}
	}()

	conn, err := db.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("error getting conn: %w", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("error closing conn: %v", err)
		}
	}()

	return conn.Raw(func(driverConn interface{}) error {
		if err := rawProcess(driverConn); err != nil {
			return fmt.Errorf("error in rawProcess: %w", err)
		}
		return nil
	})
}

func rawProcess(driverConn interface{}) error {
	conn, ok := driverConn.(*ase.Conn)
	if !ok {
		return errors.New("invalid driver, conn is not *github.com/SAP/go-ase/purego.Conn")
	}

	fmt.Println("creating table nullTypesRaw")
	if _, _, err := conn.DirectExec(context.Background(), "if object_id('nullTypesRaw') is not null drop table nullTypesRaw"); err != nil {
		return fmt.Errorf("failed to drop table 'nullTypesRaw': %w", err)
	}

	if _, _, err := conn.DirectExec(context.Background(), "create table nullTypesRaw (a BigInt null)"); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	defer func() {
		if _, _, err := conn.DirectExec(context.Background(), "drop table nullTypesRaw"); err != nil {
			log.Printf("failed to drop table: %v", err)
		}
	}()

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
		if _, _, err := conn.DirectExec(context.Background(), "insert into nullTypesRaw values (?)", val); err != nil {
			return fmt.Errorf("failed to insert values: %w", err)
		}
	}

	fmt.Println("reading table contents")
	if err := readTable(conn); err != nil {
		return fmt.Errorf("error reading table: %w", err)
	}

	return nil
}

func readTable(conn *ase.Conn) error {
	stmt, err := conn.NewStmt(context.Background(), "", "select * from nullTypesRaw", true)
	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	rows, _, err := stmt.DirectExec(context.Background())
	if err != nil {
		return fmt.Errorf("error querying with prepared statement: %w", err)
	}

	var ni sql.NullInt64

	for {
		if err := rows.Next([]driver.Value{&ni}); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("error reading row: %w", err)
		}

		value, _ := ni.Value()
		fmt.Printf("a: %v\n", value)
	}

	return nil
}
