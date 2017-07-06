# Go Drupal

[![GoDoc](https://godoc.org/github.com/phayes/go-drupal?status.svg)](https://godoc.org/github.com/phayes/go-drupal)

Package `drupal` is a go library for interacting with a drupal site via command-line and drush.

Get simple site status information:

import "github.com/phayes/go-drupal"

```go
func main() {
	site, err := drupal.NewSite("/var/www/drupalsite")
	if err != nil {
		log.Fatal(err)
	}

	status, err := site.GetStatus();
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(status.DrupalVersion)
}
```

Connect to the Drupal database:

```go
package main

import (
	"fmt"
	"log"
)

func main() {
	site, err := drupal.NewSite("/var/www/drupalsite")
	if err != nil {
		log.Fatal(err)
	}

	dbinfo, err := site.GetDatabase()
	if err != nil {
		log.Fatal(err)
	}

	db, err := dbinfo.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Test the connection to the database
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	// Do queries
	var user string
	err := db.QueryRow("SELECT name FROM users WHERE uid=?", 1).Scan(&user)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Admin username is", user)
}
```

Running a drush command:

```go
package main

import (
	"fmt"
	"log"
)

func main() {
	site, err := drupal.NewSite("/var/www/drupalsite/sites/multisite")
	if err != nil {
		log.Fatal(err)
	}

	// Run "drush cr" to rebuild cache
	output, messages, err := site.Drush("cr")
	if err != nil {
		// Optionally inspect the error to see if they are only warnings
		errset, ok := err.(drupal.DrushMessages)
		if !ok {
			log.Fatal(err) // Error running the drush command
		}
		// If it contains errors, then throw fatal error
		if errset.HasErrors() {
			log.Fatal(errset)
		}
		// If it has no errors, but has warnings, just print the warnings
		if !errset.HasErrors() && errset.HasWarnings() {
			fmt.Println(errset)
		}
		// If it doesn't contain warnings or errors, do nothing
		// This might be the case if it produces notices or other unknown stderr output
		if !errset.HasErrors() && !errset.HasWarnings() {
			// do nothing
		}
	}

	// Print out any "ok" or "success" messages
	for _, message := range messages {
		fmt.Println(message)
	}

	// Print out the output
	fmt.Println(output)
}
```

Get $settings from settings.php:

```go
package main

import (
	"fmt"
	"log"
)

func main() {
	site, err := drupal.NewSite("/var/www/drupalsite")
	if err != nil {
		log.Fatal(err)
	}

	settings, err := site.GetSettings()
	if err != nil {
		log.Fatal(err)
	}

	// Get a specific setting
	hashSalt := settings.GetString("hash_salt")

	fmt.Println("Hash Salt for site is", hashSalt)
}

```