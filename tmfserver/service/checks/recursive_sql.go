package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/hesusruiz/isbetmf/tmfserver/repository"
	_ "github.com/mattn/go-sqlite3"
)

func main() {

	// Open the sqlite file at 'data/isbetmf.db'
	db, err := sql.Open("sqlite3", "data/isbetmf.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var sql, val string

	sql = `EXISTS (SELECT 1 
FROM json_tree(tmf_object.content) 
WHERE json_tree.fullkey = '$.organizationIdentification[0].identificationId' 
  AND json_tree.value = ?)`
	val = "did:elsi:VATES-G87936159"

	// Test 0:fmt.Printf("Input: %s\n", input)
	fmt.Printf("Query 0:\n%s\nValue: %s\n", sql, val)
	runQueryAndPrintResults(db, sql, []any{val})

	vals := []any{"did:elsi:VATES-11111111K", "did:elsi:VATES-G87936159"}

	// Test 1:
	input := "organizationIdentification[0].identificationId"
	sql, _, err = repository.GenerateRecursiveJSONQuery("tmf_object", input, vals)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Input: %s\n", input)
	fmt.Printf("Query 1:\n%s\nValue: %s\n", sql, vals)
	runQueryAndPrintResults(db, sql, vals)

	// Test 2: Full wildcard search
	input = "organizationIdentification[*].identificationId"
	sql, _, err = repository.GenerateRecursiveJSONQuery("tmf_object", input, vals)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Input: %s\n", input)
	fmt.Printf("Query 2:\n%s\nValue: %s\n", sql, vals)
	runQueryAndPrintResults(db, sql, vals)

	// Test 3: Specific index search
	fmt.Printf("\n")
	input = "organizationIdentification.identificationId"
	sql, _, err = repository.GenerateRecursiveJSONQuery("tmf_object", input, vals)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Input: %s\n", input)
	fmt.Printf("Query 3:\n%s\nValue: %s\n", sql, vals)
	runQueryAndPrintResults(db, sql, vals)

}

func runQueryAndPrintResults(db *sql.DB, sql string, vals []any) {
	fmt.Println("==========")

	sqlFull := "SELECT DISTINCT tmf_object.id, tmf_object.content FROM tmf_object WHERE type='organization' AND " + sql
	fmt.Printf("Full Query:\n%s\nValue: %s\n", sqlFull, vals)
	// Run the query
	rows, err := db.Query(sqlFull, vals...)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Print the results
	for rows.Next() {
		var id string
		var content string
		err = rows.Scan(&id, &content)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("-----")
		fmt.Printf("ID: %s\n", id)
		fmt.Printf("Content: %s\n", content)
		fmt.Println("-----")
	}
	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("==========")
	fmt.Println()
}
