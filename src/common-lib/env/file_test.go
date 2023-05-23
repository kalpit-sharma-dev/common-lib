package env

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"io/ioutil"
)

var mockedExitStatus = 0
var mockedStdout, mockedStderr string

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestExecCommandHelper", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	es := strconv.Itoa(mockedExitStatus)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1",
		"STDOUT=" + mockedStdout,
		"STDERR=" + mockedStderr,
		"EXIT_STATUS=" + es}
	return cmd
}

func TestExecCommandHelper(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	fmt.Fprintf(os.Stdout, os.Getenv("STDOUT"))
	fmt.Fprintf(os.Stderr, os.Getenv("STDERR"))

	i, _ := strconv.Atoi(os.Getenv("EXIT_STATUS"))
	os.Exit(i)
}

func Test_envImpl_ExecuteBash_Success(t *testing.T) {
	mockedExitStatus = 0
	mockedStdout = "2.0"
	execCommand = fakeExecCommand
	defer func() {
		execCommand = exec.Command
	}()
	want := "2.0"

	e := envImpl{}
	got, _ := e.ExecuteBash("")
	if got != want {
		t.Errorf("Expected %q, got %q", want, got)
	}

}

func Test_envImpl_ExecuteBash_Failure_Invalid_Cmd(t *testing.T) {
	mockedExitStatus = 1
	mockedStdout = ""
	mockedStderr = "invalid cmd"
	execCommand = fakeExecCommand
	defer func() {
		execCommand = exec.Command
	}()
	want := "invalid cmd"

	e := envImpl{}
	_, err := e.ExecuteBash("")
	if err.Error() != want {
		t.Errorf("Expected %q, got %q", want, err.Error())
	}

}

func Test_envImpl_ExecuteBash_Failure_Exit_Status(t *testing.T) {
	mockedExitStatus = 1
	mockedStdout = ""
	mockedStderr = ""
	execCommand = fakeExecCommand
	defer func() {
		execCommand = exec.Command
	}()
	want := "exit status 1"

	e := envImpl{}
	_, err := e.ExecuteBash("")
	if err.Error() != want {
		t.Errorf("Expected %q, got %q", want, err.Error())
	}

}


func Test_envImpl_GetCommandReader_Success(t *testing.T) {
	mockedExitStatus = 0
	mockedStdout = "testmsg"
	execCommand = fakeExecCommand
	defer func() {
		execCommand = exec.Command
	}()
	want := "testmsg"

	e := envImpl{}
	out, _ := e.GetCommandReader("bash","-c", "echo testmsg")
	got, err := ioutil.ReadAll(out)
	if err != nil {
		t.Errorf("failed to read out; err %v", err)
	}
	if string(got) != want {
		t.Errorf("Expected %q, got %q", want, got)
	}

}

func Test_envImpl_GetCommandReader_Failure_Invalid_Cmd(t *testing.T) {
	mockedExitStatus = 1
	mockedStdout = ""
	mockedStderr = "invalid cmd"
	execCommand = fakeExecCommand
	defer func() {
		execCommand = exec.Command
	}()
	want := "invalid cmd"

	e := envImpl{}
	_, err := e.GetCommandReader("")
	if err.Error() != want {
		t.Errorf("Expected %q, got %q", want, err.Error())
	}

}

func Test_envImpl_GetCommandReader_Failure_Exit_Status(t *testing.T) {
	mockedExitStatus = 1
	mockedStdout = ""
	mockedStderr = ""
	execCommand = fakeExecCommand
	defer func() {
		execCommand = exec.Command
	}()
	want := "exit status 1"

	e := envImpl{}
	_, err := e.GetCommandReader("")
	if err.Error() != want {
		t.Errorf("Expected %q, got %q", want, err.Error())
	}

}
