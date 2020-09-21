// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

// +build integration

package main

import "log"

func ExampleDoMain() {
	if err := DoMain(); err != nil {
		log.Fatalf("nullTypes example: %v", err)
	}
	// Output:
	// creating table nullTypesRaw
	// Writing a=123 to table
	// Writing a=<nil> to table
	// Writing a=321 to table
	// reading table contents
	// a: 123
	// a: <nil>
	// a: 321
}
