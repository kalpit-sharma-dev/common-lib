package env

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	exc "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/exception"
)

var execCommand = exec.Command

//go:generate mockgen -package mock -destination=mock/mocks.go . FactoryEnv,Env

const (
	//ErrEnvUnableToGetExeDirectory error code for unable to get executable's directory
	ErrEnvUnableToGetExeDirectory = "EnvUnableToGetExeDirectory"
)

// FactoryEnv interface to return the Env
type FactoryEnv interface {
	GetEnv() Env
}

// FactoryEnvImpl returns the concrete implementation of Factory
type FactoryEnvImpl struct {
}

// GetEnv returns Env
func (FactoryEnvImpl) GetEnv() Env {
	return envImpl{}
}

// Env interface defines the methods of Env functionality
type Env interface {
	GetExeDir() (string, error)
	GetFileReader(filePath string) (io.ReadCloser, error)
	GetCommandReader(command string, arg ...string) (io.ReadCloser, error)
	GetDirectoryFileCount(dirPathExpr string, args ...[]string) (io.ReadCloser, error)
	ExecuteBash(cmd string) (string, error)

	//getFileCount(dirPath string) int
}

// envImpl retruns the concrete
type envImpl struct{}

// GetExeDir returns the executable's absolute path.
// go run will return a different path than exepcted because it keeps the exe into user's temp folder
func (envImpl) GetExeDir() (string, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", exc.New(ErrEnvUnableToGetExeDirectory, err)
	}
	return dir, nil
}

// GetFileReader returns a reader for the specified file
func (envImpl) GetFileReader(filePath string) (io.ReadCloser, error) {
	return os.Open(filePath)
}

// GetCommandReader returns a reader for the specified command
func (envImpl) GetCommandReader(command string, arg ...string) (io.ReadCloser, error) {
	cmd := execCommand(command, arg...)
	var out, errout bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errout
	err := cmd.Run()

	byteReader := bytes.NewReader(out.Bytes())
	reader := ioutil.NopCloser(byteReader)

	if errout.Len() == 0 && out.Len() == 0 {
		return reader, err
	}
	if out.Len() == 0 {
		return reader, errors.New(errout.String())
	}

	return reader, nil
}

func (envImpl) GetDirectoryFileCount(dirPathExpr string, args ...[]string) (io.ReadCloser, error) {
	var strFinalParameter []byte
	var strParameters string

	iArgumentLen := len(args)
	if iArgumentLen > 0 {
		iArrayLen := len(args[0])
		strInterfaceArr := make([]interface{}, iArgumentLen)

		for iCtr := 0; iCtr < iArrayLen; iCtr++ {
			for iSubCtr := 0; iSubCtr < iArgumentLen; iSubCtr++ {
				strInterfaceArr[iSubCtr] = args[iSubCtr][iCtr]
			}

			dPath := fmt.Sprintf(dirPathExpr, strInterfaceArr...)
			fileCount := getFileCount(dPath)

			strParameters = fmt.Sprintf("%s ", dPath)
			for iIndex := 0; iIndex < iArgumentLen; iIndex++ {
				strParameters += fmt.Sprint(strInterfaceArr[iIndex])
				strParameters += " "
			}
			strParameters += fmt.Sprintf("%d\n", fileCount)

			strFinalParameter = append(strFinalParameter, strParameters...)
		}
	} else {
		fileCount := getFileCount(dirPathExpr)
		strParameters = fmt.Sprintf("%s ", dirPathExpr)
		strParameters += fmt.Sprintf("%d\n", fileCount)

		strFinalParameter = append(strFinalParameter, strParameters...)
	}

	objReadCloser := ioutil.NopCloser(bytes.NewReader(strFinalParameter))
	return objReadCloser, nil
}

// ExecuteBash is to execute bash command and return it's output
func (envImpl) ExecuteBash(cmd string) (string, error) {
	command := execCommand("bash", "-c", cmd)
	var out, errout bytes.Buffer
	command.Stdout = &out
	command.Stderr = &errout
	err := command.Run()
	if errout.Len() == 0 && out.Len() == 0 {
		return "", err
	}
	if out.Len() == 0 {
		return "", errors.New(errout.String())
	}
	return strings.TrimSpace(out.String()), nil
}

func getFileCount(dirPath string) int {
	fdir, err := os.Open(dirPath)
	if nil != err {
		return 0
	}
	fileArr, err := fdir.Readdir(-1)
	if nil != err {
		fdir.Close() //nolint
		return 0
	}
	fdir.Close() //nolint
	return len(fileArr)
}
