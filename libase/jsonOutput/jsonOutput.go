// Package should be as generic as possible since it could be used from other sql.driver too
package jsonOutput

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

type results struct {
	Query string
	Rows  []row
}

type row map[string]interface{}

// get second query and column colA
// pgo -json "use master; select * from tt3;" | jq ' [ .[1].Rows[].colA ]'

func Process(ctx context.Context, db *sql.DB, parsedQueries []string) error {

	var resultSet []results

	// open exclusive connection
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("jsonOutput: failed to open connection: %w", err)
	}
	defer conn.Close()

	// Loop over queries and execute
	for _, query := range parsedQueries {
		results, err := process(ctx, conn, query)
		if err != nil {
			return fmt.Errorf("jsonOutput: failed to process query '%s': %w", query, err)
		}

		// append result to resultSet
		resultSet = append(resultSet, results)
	}

	// output
	if err := jsonOutput(resultSet); err != nil {
		return fmt.Errorf("jsonOutput: failed to process jsonOutput: %w", err)
	}

	return nil
}

// ExecQuery executes an array of given queries and store the results in
// resultSet
func process(ctx context.Context, conn *sql.Conn, query string) (results, error) {

	var results results
	results.Query = query

	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		return results, fmt.Errorf("jsonOutput: failed to queryContext %s: %w", query, err)
	}
	defer rows.Close()

	for rows.Next() {
		row := make(map[string]interface{})

		// get columns
		columns, err := rows.Columns()
		if err != nil {
			return results, fmt.Errorf("jsonOutput: failed to get columns: %w", err)
		}

		values := make([]interface{}, len(columns))
		// init value-pointer
		for i := range columns {
			var value interface{}
			values[i] = &value
		}

		// scan rows
		if err := rows.Scan(values...); err != nil {
			return results, fmt.Errorf("jsonOutput: failed to scan rows: %w", err)
		}

		// insert columns and values to row
		for i, value := range values {
			row[columns[i]] = value
		}

		// append row to struct
		results.Rows = append(results.Rows, row)
	}

	return results, nil
}

func jsonOutput(resultSet []results) error {
	output, err := json.MarshalIndent(resultSet, "", "    ")
	if err != nil {
		return fmt.Errorf("jsonOutput: failed to MarshalIndent resultSet %v: %w", resultSet, err)
	}
	// print output
	fmt.Println(string(output))
	return nil
}
