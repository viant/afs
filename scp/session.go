package scp

import (
	"bytes"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/viant/afs/file"
	"github.com/viant/afs/storage"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

const (
	modeWrite = iota
	modeRead
	statusOK         = 0x0
	defaultTimeoutMs = 15000
)

type session struct {
	*ssh.Session
	skipBaseDir  bool
	locationName string
	counter      uint32
	reader       *reader
	writer       io.WriteCloser
	mode         int
	timeout      time.Duration
	recursive    bool
}

func (s *session) initCmd(location string) string {
	_, s.locationName = path.Split(location)
	var cmdSwitches = make([]string, 0)
	switch s.mode {
	case modeWrite:
		cmdSwitches = append(cmdSwitches, "-t -p")
	default:
		cmdSwitches = append(cmdSwitches, "-f -p")
	}
	if s.recursive {
		cmdSwitches = append(cmdSwitches, "-r")
	}
	return fmt.Sprintf("scp %v %v\n", strings.Join(cmdSwitches, " "), location)
}

func (s *session) writeStatusOK() error {
	return s.write([]byte{statusOK})
}

func (s *session) runCmd(data string) error {
	err := s.write([]byte(data))
	if err == nil {
		err = s.readStatus()
	}
	return err
}

func (s *session) sendToken(token byte) error {
	err := s.write([]byte{token, '\n'})
	if err == nil {
		err = s.readStatus()
	}
	return err
}

func (s *session) write(data []byte) error {
	_, err := s.writer.Write(data)
	return err
}

func (s *session) init(location string) (err error) {
	if strings.HasSuffix(location, "/") {
		location = string(location[:len(location)-1])
	}
	if s.writer, err = s.StdinPipe(); err != nil {
		return err
	}
	reader, err := s.StdoutPipe()
	if err != nil {
		return err
	}
	s.reader = newReader(reader)
	cmd := s.initCmd(location)
	if err = s.Start(cmd); err != nil {
		return err
	}
	go s.reader.readInBackground()
	return err
}

func (s *session) readContent(info os.FileInfo) (io.Reader, error) {
	buf := new(bytes.Buffer)
	for buf.Len() <= int(info.Size()) {
		data, err := s.reader.read(s.timeout)
		if err != nil {
			return nil, err
		}
		buf.Write(data)

	}
	data := buf.Bytes()
	overflow := data[info.Size():]
	if len(overflow) != 1 || overflow[0] != statusOK {
		return nil, fmt.Errorf("invalid statusOK, expected: %v, but had: %v ", statusOK, overflow)
	}
	data = data[:info.Size()]
	return bytes.NewReader(data), nil
}

func (s *session) processNewResource(relativeElements *[]string, response []byte, modified *time.Time, handler func(parent string, info os.FileInfo, reader io.Reader) (bool, error)) (bool, error) {

	fileInfo, err := NewInfo(string(response), modified)
	if err != nil {
		return false, err
	}
	var reader io.Reader
	parent := path.Join(*relativeElements...)
	if fileInfo.IsDir() {
		if s.skipBaseDir && s.counter == 0 && fileInfo.Name() == s.locationName {
			return false, nil
		}
		*relativeElements = append(*relativeElements, fileInfo.Name())
	} else if err = s.writeStatusOK(); err == nil {
		reader, err = s.readContent(fileInfo)
	}
	if err != nil {
		return false, err
	}
	toContinue, err := handler(parent, fileInfo, reader)
	return toContinue, err
}

func (s *session) download(ctx context.Context, skipBaseDir bool, location string, handler func(parent string, info os.FileInfo, reader io.Reader) (bool, error)) error {
	if s.mode == modeWrite {
		return fmt.Errorf("invalid mode")
	}
	err := s.init(location)
	if err != nil {
		return errors.Wrap(err, "failed to initialise session")
	}
	s.skipBaseDir = skipBaseDir
	now := time.Now()
	modified := &now

	var pathElements = make([]string, 0)
	for {
		err = s.pull(&pathElements, modified, handler)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			break
		}
	}

	return err
}

func (s *session) pull(pathElements *[]string, modified *time.Time, handler func(parent string, info os.FileInfo, reader io.Reader) (bool, error)) error {
	err := s.writeStatusOK()
	if err != nil {
		return err
	}
	response, err := s.reader.read(s.timeout)
	if err != nil {
		return err
	}
	token := response[0]
	switch token {
	case FileToken, DirToken:

		shallContinue, err := s.processNewResource(pathElements, response, modified, handler)
		s.counter++
		if err != nil || !shallContinue {
			return err
		}

	case EndDirToken:
		if len(*pathElements) > 0 {
			*pathElements = (*pathElements)[:len(*pathElements)-1]
		}
	case TimestampToken:
		timestamp, err := ParseTimeResponse(string(response))
		if err != nil {
			return err
		}
		*modified = *timestamp
	case WarningToken, ErrorToken:
		errorMessage := strings.TrimSpace(string(response[1:]))
		return fmt.Errorf("%s", errorMessage)
	default:
		return fmt.Errorf("unsupported token: %v, %s", token, response)
	}
	return nil
}

func (s *session) readStatus() error {
	data, err := s.reader.read(s.timeout)
	if err != nil {
		return err
	}
	status := data[0]
	switch status {
	case statusOK:
		return nil
	default:
		return errors.New(strings.TrimSpace(string(data[1:])))
	}
}

func (s *session) moveUp() error {
	return s.sendToken(EndDirToken)
}

func (s *session) moveDown(info os.FileInfo) error {
	return s.push(info, nil)
}

func (s *session) upload(location string) (storage.Upload, io.Closer, error) {
	if s.mode == modeRead {
		return nil, nil, fmt.Errorf("invalid mode")
	}
	err := s.init(location)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to initialise session")
	}

	err = s.readStatus()
	if err != nil {
		return nil, nil, err
	}

	var prevRelativeElements = make([]string, 0)
	handler := func(ctx context.Context, parent string, info os.FileInfo, reader io.Reader) error {
		prevRelativePath := path.Join(prevRelativeElements...)
		parent = strings.Trim(parent, "/")
		prevRelativeElements = []string{}
		if parent != "" {
			prevRelativeElements = strings.Split(parent, "/")
		}
		err = adjustPath(prevRelativePath, parent, s.moveDown, s.moveUp)

		if info.IsDir() {
			prevRelativeElements = append(prevRelativeElements, info.Name())
			return s.moveDown(info)
		}
		return s.push(info, reader)
	}
	return handler, s, nil

}

func (s *session) push(info os.FileInfo, reader io.Reader) error {
	if info.Mode()&os.ModeSymlink > 0 {
		//update size for symlink, otherwise size may be reported incorrectly
		data, _ := ioutil.ReadAll(reader)
		info = file.NewInfo(info.Name(), int64(len(data)), info.Mode().Perm(), info.ModTime(), info.IsDir())
		reader = bytes.NewReader(data)
	}

	timestampCmd := InfoToTimestampCmd(info)
	err := s.runCmd(timestampCmd)
	if err == nil {
		createCmd := InfoToCreateCmd(info)
		err = s.runCmd(createCmd)
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		if _, err = io.Copy(s.writer, reader); err != nil {
			return err
		}
		err = s.writeStatusOK()
		if err == nil {
			err = s.readStatus()
		}
	}
	return err
}

func (s *session) Close() error {
	s.reader.sendCloseNotification()
	s.reader.close()
	return s.Session.Close()
}

func newSession(client *ssh.Client, mode int, recursive bool, timeout time.Duration) (*session, error) {
	sshSession, err := client.NewSession()
	if err != nil {
		return nil, err
	}

	return &session{Session: sshSession,
		mode:      mode,
		timeout:   timeout,
		recursive: recursive,
	}, nil
}
