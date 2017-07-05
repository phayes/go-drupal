package drupal

import (
	"bufio"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/phayes/errors"
)

type Drush struct {
	Directory string
	Command   string
	Arguments []string
	cmd       *exec.Cmd
}

func NewDrush(directory string, command string, arguments ...string) *Drush {
	drush := Drush{Directory: directory, Command: command, Arguments: arguments}

	global := []string{command, "--yes", "--nocolor"}
	arguments = append(global, arguments...)

	drush.cmd = exec.Command("drush", arguments...)
	drush.cmd.Dir = drush.Directory
	drush.cmd.Env = append(os.Environ(), "DRUSH_COLUMNS=10000", "COLUMNS=10000")

	return &drush
}

func (d *Drush) Run() (output []string, warns error, errs error) {
	stderr, err := d.cmd.StderrPipe()
	if err != nil {
		return nil, nil, err
	}
	stdout, err := d.cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}

	err = d.cmd.Start()
	if err != nil {
		return nil, nil, err
	}

	// Wait Groups
	var wg sync.WaitGroup

	// Stderr
	errset := errors.NewErrorSet()
	warnset := errors.NewErrorSet()
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		i := 0
		for scanner.Scan() {
			i++
			line := scanner.Text()
			if strings.HasSuffix(line, "[error]") {
				line = strings.TrimSuffix(line, "[error]")
				line = strings.TrimSpace(line)
				errset.Add(strconv.Itoa(i), errors.New(line))
			}
			if strings.HasSuffix(line, "[warning]") {
				line = strings.TrimSuffix(line, "[warning]")
				line = strings.TrimSpace(line)
				warnset.Add(strconv.Itoa(i), errors.New(line))
			}
			if strings.HasSuffix(line, "[notice]") {
				line = strings.TrimSuffix(line, "[notice]")
				line = strings.TrimSpace(line)
				warnset.Add(strconv.Itoa(i), errors.New(line))
			}
		}
	}()

	// Stdout
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			output = append(output, strings.TrimSpace(line))
		}
	}()

	err = d.cmd.Wait()
	if err != nil {
		errset.Add("drush", err)
	}
	wg.Wait()

	// Report warnings and errors
	if warnset.HasErrors() {
		warns = warnset
	}
	if errset.HasErrors() {
		errs = errset
	}

	return
}
