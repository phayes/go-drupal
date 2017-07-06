package drupal

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type Drush struct {
	Directory string
	Command   string
	Arguments []string
	cmd       *exec.Cmd
}

func NewDrush(directory string, command string, arguments ...string) *Drush {
	drush := Drush{Directory: directory, Command: command, Arguments: arguments}
	return &drush
}

func (d *Drush) Run() (output string, messages DrushMessageSet, errs error) {
	d.buildCommand()

	stderr, err := d.cmd.StderrPipe()
	if err != nil {
		return "", nil, err
	}
	stdout, err := d.cmd.StdoutPipe()
	if err != nil {
		return "", nil, err
	}

	err = d.cmd.Start()
	if err != nil {
		return "", nil, err
	}

	// Wait Groups
	var wg sync.WaitGroup

	// Stderr
	errset := DrushMessageSet{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			message := NewDrushMessage(scanner.Text())
			if message.Type == DrushMessageOK {
				messages = append(messages, message)
			} else {
				errset = append(errset, message)
			}
		}
	}()

	// Stdout
	wg.Add(1)
	outbuf := new(bytes.Buffer)
	go func() {
		defer wg.Done()
		outbuf.ReadFrom(stdout)
	}()

	err = d.cmd.Wait()
	if err != nil {
		errset = append(errset, NewDrushMessage(err.Error()))
	}
	wg.Wait()

	if errset.HasErrors() {
		errs = errset
	}

	return outbuf.String(), messages, errs
}

func (d *Drush) buildCommand() {
	global := []string{d.Command, "--yes", "--nocolor"}
	arguments := append(global, d.Arguments...)

	d.cmd = exec.Command("drush", arguments...)
	d.cmd.Dir = d.Directory
	d.cmd.Env = append(os.Environ(), "DRUSH_COLUMNS=10000", "COLUMNS=10000")
}

// DrushMessage implements the standard error interface and represents a single line in stdout
type DrushMessage struct {
	Message string
	Type    DrushMessageType
}

// NewDrushMessage returns a DrushMessage from a raw stderr output line from a drush command
func NewDrushMessage(messageline string) DrushMessage {
	var mestype DrushMessageType
	if strings.HasSuffix(messageline, DrushMessageError.String()) {
		mestype = DrushMessageError
	} else if strings.HasSuffix(messageline, DrushMessageWarning.String()) {
		mestype = DrushMessageWarning
	} else if strings.HasSuffix(messageline, DrushMessageNotice.String()) {
		mestype = DrushMessageWarning
	} else if strings.HasSuffix(messageline, DrushMessageOK.String()) {
		mestype = DrushMessageOK
	} else {
		mestype = DrushMessageUnknown
	}

	messageline = strings.TrimSuffix(messageline, mestype.String())
	messageline = strings.TrimSpace(messageline)

	return DrushMessage{Message: messageline, Type: mestype}
}

func (de DrushMessage) Error() string {
	return de.Type.String() + ": " + de.Message
}

// DrushMessageType specifies the type of error (eg error, warnings, notice, unknown)
type DrushMessageType string

func (det DrushMessageType) String() string {
	return string(det)
}

const (
	DrushMessageError   DrushMessageType = "[error]"   // For errors reported as [error]
	DrushMessageWarning DrushMessageType = "[warning]" // For errors reported as [warning]
	DrushMessageNotice  DrushMessageType = "[notice]"  // For errors reported as [notice]
	DrushMessageOK      DrushMessageType = "[ok]"      // For non-errors reported as [ok]
	DrushMessageUnknown DrushMessageType = "[unknown]" // All other output in stderr
)

// DrushMessageSet implements the standard error interface and represents all errors, warnings and notices reported by a drush command
type DrushMessageSet []DrushMessage

func (des DrushMessageSet) Error() string {
	if des == nil {
		return ""
	}
	output := ""

	for _, DrushMessage := range des {
		output += DrushMessage.Error() + ". "
	}

	return output
}

// HasErrors checks to see if the DrushMessageSet contains [error] errors.
// It will return false if the DrushMessageSet only contains warnings and notices.
func (des DrushMessageSet) HasErrors() bool {
	if des == nil {
		return false
	}

	for _, DrushMessage := range des {
		if DrushMessage.Type == DrushMessageError {
			return true
		}
	}
	return false
}

// HasWarnings checks to see if the DrushMessageSet contains [warning] errors.
// It will return false if the DrushMessageSet only contains errors and notices.
func (des DrushMessageSet) HasWarnings() bool {
	if des == nil {
		return false
	}

	for _, DrushMessage := range des {
		if DrushMessage.Type == DrushMessageWarning {
			return true
		}
	}
	return false
}

// HasNotices checks to see if the DrushMessageSet contains [notice] errors.
// It will return false if the DrushMessageSet only contains errors and warnings.
func (des DrushMessageSet) HasNotices() bool {
	if des == nil {
		return false
	}

	for _, DrushMessage := range des {
		if DrushMessage.Type == DrushMessageNotice {
			return true
		}
	}
	return false
}

// HasUnknowns checks to see if the DrushMessageSet contains unknown errors in stderr.
func (des DrushMessageSet) HasUnknowns() bool {
	if des == nil {
		return false
	}

	for _, DrushMessage := range des {
		if DrushMessage.Type == DrushMessageUnknown {
			return true
		}
	}
	return false
}
