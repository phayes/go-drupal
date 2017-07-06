package drupal

import (
	"database/sql"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/phayes/errors"
)

// Site represents a Drupal site, defined by it's location in the filesystem
type Site string

// NewSite returns a Site, given a directory
func NewSite(rootDirectory string) (Site, error) {
	var err error

	rootDirectory, err = filepath.Abs(rootDirectory)
	if err != nil {
		return "", errors.Wrapf(err, "Drupal site error. Could not determine absolute path of %v", rootDirectory)
	}

	info, err := os.Stat(rootDirectory)
	if err != nil {
		return "", errors.Wrapf(err, "Drupal site error. Could not stat %v", rootDirectory)
	}
	if !info.IsDir() {
		return "", errors.Newf("Drupal site error. %v is not a directory", rootDirectory)
	}

	_, err = exec.LookPath("php")
	if err != nil {
		return "", errors.Wraps(err, "Drupal site error. php executable not found")
	}

	return Site(rootDirectory), nil
}

// GetSettings gets the $settings array defined in settings.php
func (s Site) GetSettings() (Settings, error) {
	status, err := s.GetStatus()
	if err != nil {
		return nil, err
	}

	phpCode := "$app_root = '" + status.Root + "'; $site_path = '" + status.Site + "'; include_once($app_root.'/'.$site_path.'/settings.php'); print json_encode($settings);"

	out, err := exec.Command("php", "-r", phpCode).Output()
	if err != nil {
		return nil, errors.Wraps(err, "Error fetching drupal settings")
	}

	var settings Settings
	err = json.Unmarshal(out, &settings)
	if err != nil {
		return nil, errors.Wraps(err, "Error fetching drupal settings")
	}
	return settings, nil
}

// GetStatus gets the Status from "drush status"
func (s Site) GetStatus() (*Status, error) {
	output, _, errs := s.Drush("status", "--format=json")
	if errs != nil {
		return nil, errs
	}

	status := &Status{}
	err := json.Unmarshal([]byte(output), status)
	if err != nil {
		return nil, err
	}

	return status, nil
}

// GetDefaultDatabase returns the database connection details for the default database connection
func (s Site) GetDefaultDatabase() (*Database, error) {
	status, err := s.GetStatus()
	if err != nil {
		return nil, err
	}

	phpCode := "$app_root = '" + status.Root + "'; $site_path = '" + status.Site + "'; include_once($app_root.'/'.$site_path.'/settings.php'); print json_encode($databases['default']['default']);"

	out, err := exec.Command("php", "-r", phpCode).Output()
	if err != nil {
		return nil, errors.Wraps(err, "Error fetching drupal database")
	}

	var defaultDatabase Database
	err = json.Unmarshal(out, &defaultDatabase)
	if err != nil {
		return nil, errors.Wraps(err, "Error fetching drupal database")
	}

	return &defaultDatabase, nil
}

// String returns the directory for the drupal site
func (s Site) String() string {
	return string(s)
}

// Drush runs a drush command.
//
// Example:
//	import "github.com/phayes/go-drupal"
//
//	func main() {
//		site, err := drupal.NewSite("/var/www/drupalsite/sites/multisite")
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		// Run "drush cr" to rebuild cache
//		output, messages, err := site.Drush("cr")
//		if err != nil {
//			// Optionally inspect the error to see if they are only warnings
//			errset, ok := err.(drupal.DrushMessages)
//			if !ok {
//				log.Fatal(err) // Error running the drush command
//			}
//			// If it contains errors, then throw fatal error
//			if errset.HasErrors() {
//				log.Fatal(errset)
//			}
//			// If it has no errors, but has warnings, just print the warnings
//			if !errset.HasErrors() && errset.HasWarnings() {
//				fmt.Println(errset)
//			}
//			// If it doesn't contain warnings or errors, do nothing
//			// This might be the case if it produces notices or other unknown stderr output
//			if !errset.HasErrors() && !errset.HasWarnings() {
//				// do nothing
//			}
//		}
//
//		// Print out any "ok" or "success" messages
//		for _, message := range messages {
//			fmt.Println(message)
//		}
//
//		// Print out the output
//		fmt.Println(output)
//	}
func (s Site) Drush(command string, arguments ...string) (output string, messages DrushMessages, errs error) {
	drush := NewDrush(s.String(), command, arguments...)
	return drush.Run()
}

// Database represents database connection details for a drupal site
type Database struct {
	Database  string `json:"database"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Prefix    string `json:"prefix"`
	Host      string `json:"host"`
	Port      string `json:"port"`
	Namespace string `json:"namespace"`
	Driver    string `json:"driver"`
}

// Open opens a connection to the database
// Be sure to call "Close()" on the provided connection when done
func (db *Database) Open() (*sql.DB, error) {
	// Create an sql.DB and check for errors
	connection := db.Username
	if db.Password != "" {
		connection += ":" + db.Password
	}
	connection += "@" + db.Host
	if db.Port != "" {
		connection += ":" + db.Port
	}
	connection += "/" + db.Database

	return sql.Open(db.Driver, connection)
}

// Status contain miscalaneous information about a drupal site, obtained from "drush status"
type Status struct {
	DrupalVersion      string   `json:"drupal-version"`
	DrupalSettingsFile string   `json:"drupal-settings-file"`
	URI                string   `json:"uri"`
	DBDriver           string   `json:"db-driver"`
	DBHostname         string   `json:"db-hostname"`
	DBUsername         string   `json:"db-username"`
	DBName             string   `json:"db-name"`
	DBPort             string   `json:"db-port"`
	PHPBin             string   `json:"php-bin"`
	PHPOS              string   `json:"php-os"`
	PHPConf            []string `json:"php-conf"`
	DrushScript        string   `json:"drush-script"`
	DrushVersion       string   `json:"drush-version"`
	DrushTemp          string   `json:"drush-temp"`
	DrushConf          []string `json:"drush-conf"`
	DrushAliasFiles    []string `json:"drush-alias-files"`
	Root               string   `json:"root"`
	Site               string   `json:"site"`
	Modules            string   `json:"modules"`
	Themes             string   `json:"themes"`
	ConfigSync         string   `json:"config-sync"`
}
