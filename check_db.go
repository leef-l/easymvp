package main

import (
	"database/sql"
	"fmt"
	"os"
	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", `C:\Users\Public\project\easymvp\var\data\easymvp.db`)
	if err != nil {
		fmt.Println("open error:", err)
		os.Exit(1)
	}
	defer db.Close()

	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table'")
	if err != nil {
		fmt.Println("query error:", err)
		os.Exit(1)
	}
	defer rows.Close()

	fmt.Println("Tables:")
	count := 0
	for rows.Next() {
		var name string
		rows.Scan(&name)
		fmt.Println("  ", name)
		count++
	}
	if count == 0 {
		fmt.Println("  (none)")
	}
}
