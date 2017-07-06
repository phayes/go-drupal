package drupal

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/phayes/errors"
)

type Site struct {
	Directory string
}

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

func NewSite(rootDirectory string) (*Site, error) {
	var err error

	rootDirectory, err = filepath.Abs(rootDirectory)
	if err != nil {
		return nil, errors.Wrapf(err, "Drupal site error. Could not determine absolute path of %v", rootDirectory)
	}

	info, err := os.Stat(rootDirectory)
	if err != nil {
		return nil, errors.Wrapf(err, "Drupal site error. Could not stat %v", rootDirectory)
	}
	if !info.IsDir() {
		return nil, errors.Newf("Drupal site error. %v is not a directory", rootDirectory)
	}

	_, err = exec.LookPath("php")
	if err != nil {
		return nil, errors.Wraps(err, "Drupal site error. php executable not found")
	}

	return &Site{Directory: rootDirectory}, nil
}

func (s *Site) GetSettings() (Settings, error) {
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

func (s *Site) GetStatus() (*Status, error) {
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

func (s *Site) GetDefaultDatabase() (*Database, error) {
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

func (s *Site) Drush(command string, arguments ...string) (output string, messages DrushMessageSet, errs error) {
	drush := NewDrush(s.Directory, command, arguments...)
	return drush.Run()
}
