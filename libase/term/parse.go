// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package term

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/SAP/go-ase/libase/jsonOutput"
	"strings"
)

func ParseAndExecQueries(db *sql.DB, line string) error {
	builder := strings.Builder{}
	currentlyQuoted := false

	var queries []string

	for i, chr := range line {
		switch chr {
		case '"', '\'':
			if currentlyQuoted {
				currentlyQuoted = false
				builder.WriteRune(chr)
			} else {
				currentlyQuoted = true
				builder.WriteRune(chr)
			}
		case ';':
			if currentlyQuoted {
				builder.WriteRune(chr)
			} else {
				if *fJson {
					// If builder contains query -> append parsed queries
					if builder.Len() != 0 {
						query := strings.TrimLeft(builder.String(), " ")
						queries = append(queries, query)
					}
					// if line ended, give 'queries' to ExecQuery
					if i == (len(line) - 1) {
						if err := jsonOutput.Process(context.Background(), db, queries); err != nil {
							return fmt.Errorf("term: failed to process query to json: %w", err)
						}
					}
				} else {
					if err := process(db, builder.String()); err != nil {
						return fmt.Errorf("term: failed to process query: %w", err)
					}
				}
				builder.Reset()
			}
		default:
			builder.WriteRune(chr)
		}
	}

	if builder.String() != "" {
		if err := process(db, builder.String()); err != nil {
			return fmt.Errorf("term: failed to process query: %w", err)
		}
	}

	return nil
}
