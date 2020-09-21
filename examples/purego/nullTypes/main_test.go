// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

// +build integration

package main

import "log"

func ExampleDoMain() {
	if err := DoMain(); err != nil {
		log.Printf("Failed to execute example: %v", err)
	}
	// Output:
	//
	// Opening database
	// Creating table 'nullTypes'
	// Writing a=123 to table
	// Writing a=<nil> to table
	// Writing a=321 to table
	// Querying values from table
	// Displaying results of query
	// | a          |
	// | 123        |
	// | <nil>      |
	// | 321        |
}
