package drupal

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// Drush is a drush command to be executed
type Drush struct {
	Directory string
	Command   string
	Arguments []string
	cmd       *exec.Cmd
}

// NewDrush returns a new drush command
func NewDrush(directory string, command string, arguments ...string) *Drush {
	drush := Drush{Directory: directory, Command: command, Arguments: arguments}
	return &drush
}

// Run executes the drush command
// output is the output written to stdout
// messages are any [ok] or [success] messages written to stderr
// errs might be an instance of DrushMessages, and will contain errors, warnings, and notices produced by the command
// To inspect individual errors do the following:
//   output, messages, errs := myDrushCommand.Run()
//   if errs != nil {
//     errset, ok := errs.(DrushMessages)
//     if !ok {
//       return err // Not an instances of DrushMessages, command likely failed to start
//     }
//     // Inspect individual errors, warnings and notices
//     for _, message := range errset {
//       …
//     }
//   }
func (d *Drush) Run() (output string, messages DrushMessages, errs error) {
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
	errset := DrushMessages{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			message := NewDrushMessage(scanner.Text())
			if message.Type == DrushMessageOK || message.Type == DrushMessageSuccess {
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

	if errset != nil && len(errset) > 0 {
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
	} else if strings.HasSuffix(messageline, DrushMessageSuccess.String()) {
		mestype = DrushMessageSuccess
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
	DrushMessageSuccess DrushMessageType = "[success]" // For non-errors reported as [success]
	DrushMessageUnknown DrushMessageType = "[unknown]" // All other output in stderr
)

// DrushMessages implements the standard error interface and represents all errors, warnings and notices reported by a drush command
type DrushMessages []DrushMessage

func (des DrushMessages) Error() string {
	if des == nil {
		return ""
	}
	output := ""

	for _, DrushMessage := range des {
		output += DrushMessage.Error() + " "
	}

	return output
}

// HasErrors checks to see if the DrushMessages contains [error] errors.
// It will return false if the DrushMessages only contains warnings and notices.
func (des DrushMessages) HasErrors() bool {
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

// HasWarnings checks to see if the DrushMessages contains [warning] errors.
// It will return false if the DrushMessages only contains errors and notices.
func (des DrushMessages) HasWarnings() bool {
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

// HasNotices checks to see if the DrushMessages contains [notice] errors.
// It will return false if the DrushMessages only contains errors and warnings.
func (des DrushMessages) HasNotices() bool {
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

// HasUnknowns checks to see if the DrushMessages contains unknown errors in stderr.
func (des DrushMessages) HasUnknowns() bool {
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
