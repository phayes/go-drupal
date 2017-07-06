package drupal

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDownload(t *testing.T) {
	site, err := NewSite("./test")
	if err != nil {
		t.Error(err)
	}

	// Download drupal core if needed
	if _, err := os.Stat("./test/drupal-8.3.5"); os.IsNotExist(err) {
		fmt.Println("Downloading drupal core")
		site.Drush("dl", "drupal-8.3.5")
	}

	// Set the site directory to the downloaded drupal core
	site = "./test/drupal-8.3.5"

	// Symlink settings.php
	source, err := filepath.Abs("./test/settings.php")
	if err != nil {
		t.Error(err)
	}
	target, err := filepath.Abs("./test/drupal-8.3.5/sites/default/settings.php")
	if err != nil {
		t.Error(err)
	}
	os.Symlink(source, target)

	// Test download of views
	output, messages, errs := site.Drush("dl", "views")
	if output == "" {
		t.Error("Empty output on drush dl views")
	}
	if messages[0].Type != DrushMessageSuccess {
		t.Error("No success message on drush dl views")
	}
	if errs == nil {
		t.Error("No warning messages on drush dl views")
	}
	errMessages, ok := errs.(DrushMessages)

	if !ok {
		t.Error("Could not transform errs to DrushMessages")
		return
	}

	if len(errMessages) != 1 {
		t.Error("Incorrect number of warnings on drush dl views")
	}
	if errMessages[0].Type != DrushMessageWarning {
		t.Error("Wrong type of error on drush dl views")
	}
	if errMessages[0].Message != "There are no stable releases for project views." {
		t.Error("Wrong error message on drush dl views. Got", errMessages[0].Message)
	}
}

func TestStatus(t *testing.T) {
	site, err := NewSite("./test/drupal-8.3.5")
	if err != nil {
		t.Error(err)
	}

	status, err := site.GetStatus()
	if err != nil {
		t.Error(err)
	}

	if status.DrupalVersion != "8.3.5" {
		t.Error("Bad status.DrupalVersion")
	}
	if status.DrupalSettingsFile != "sites/default/settings.php" {
		t.Error("Bad status.DrupalSettingsFile")
	}
	if status.DBDriver != "mysql" {
		t.Error("Bad status.DBDriver")
	}
	if status.DBHostname != "mysql" {
		t.Error("Bad status.DBHost")
	}
	if status.Site != "sites/default" {
		t.Error("Bad status.Site")
	}

}

func TestSettings(t *testing.T) {
	site, err := NewSite("./test/drupal-8.3.5")

	if err != nil {
		t.Error(err)
	}

	settings, err := site.GetSettings()
	if err != nil {
		t.Error(err)
	}

	fileScanDirs := settings.GetArray("file_scan_ignore_directories")

	if !reflect.DeepEqual(fileScanDirs, []string{"node_modules", "bower_components"}) {
		t.Error("Bad length for fileScanDirs")
	}

	hashSalt := settings.GetString("hash_salt")
	if hashSalt != "HASH SALT TEST" {
		t.Error("Bad hash salt")
	}
}

func TestDatabase(t *testing.T) {
	site, err := NewSite("./test/drupal-8.3.5")

	if err != nil {
		t.Error(err)
	}

	database, err := site.GetDefaultDatabase()
	if err != nil {
		t.Error(err)
	}

	if database.Database != "drupal" {
		t.Error("Bad database database")
	}
	if database.Username != "root" {
		t.Error("Bad database username")
	}
	if database.Password != "" {
		t.Error("Bad database password")
	}
	if database.Prefix != "" {
		t.Error("Bad database prefix")
	}
	if database.Host != "mysql" {
		t.Error("Bad database host")
	}
	if database.Port != "3306" {
		t.Error("Bad database port")
	}
	if database.Driver != "mysql" {
		t.Error("Bad database driver")
	}
}

func TestDrush(t *testing.T) {

	// Test Status command
	site, err := NewSite("./test/drupal-8.3.5")

	if err != nil {
		t.Error(err)
	}

	output, info, errs := site.Drush("status")
	if info != nil {
		t.Error("Got info on drush status")
	}
	if errs != nil {
		t.Error("Got error on drush status")
	}
	if len(output) == 0 {
		t.Error("Got empty output on drush status")
	}

	// Test failing command
	drush := NewDrush("./test", "pm-list")
	output, info, errs = drush.Run()
	if errs != nil {
		errset, ok := errs.(DrushMessages)
		if !ok {
			// Would normally return error here
			t.Error("Got single error, not a set of errors")
		}
		if len(errset) != 3 {
			t.Error("Incorrect number of errors")
		}
	}

	// Check output
	if output != "" {
		t.Error("Bad output on failing command")
	}
	if info != nil {
		t.Error("Bad info on failing command")
	}
	if errs == nil {
		t.Error("Errors should not be nil on failing command")
	}

}
