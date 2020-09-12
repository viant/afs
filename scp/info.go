package scp

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/viant/afs/file"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultFileMode os.FileMode = 0755
)

//NewInfo returns new info from SCP response
func NewInfo(createResponse string, modified *time.Time) (os.FileInfo, error) {

	elements := strings.SplitN(createResponse, " ", 3)
	if len(elements) != 3 {
		return nil, fmt.Errorf("invalid download createResponse: %v", createResponse)
	}
	isDir := strings.HasPrefix(elements[0], "D")
	modeLiteral := string(elements[0][1:])
	mode, err := strconv.ParseInt(modeLiteral, 8, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid mode: %v", modeLiteral)
	}
	sizeLiteral := elements[1]
	size, err := strconv.ParseInt(sizeLiteral, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid size: %v", modeLiteral)
	}
	name := strings.Trim(elements[2], "\r\n")
	if modified == nil {
		now := time.Now()
		modified = &now
	}
	return file.NewInfo(name, size, os.FileMode(mode), *modified, isDir), nil
}

//ParseTimeResponse parases respons time
func ParseTimeResponse(response string) (*time.Time, error) {
	elements := strings.SplitN(response, " ", 4)
	if len(elements) != 4 {
		return nil, fmt.Errorf("invalid timestamp response: %v", response)
	}
	unixTimestampLiteral := elements[0][1:]
	unixTimestamp, err := strconv.ParseInt(unixTimestampLiteral, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid timestamp: %v", unixTimestampLiteral)
	}
	msecLiteral := elements[1]
	msec, _ := strconv.ParseInt(msecLiteral, 10, 64)
	ts := time.Unix(unixTimestamp, msec*1000)
	return &ts, nil
}

//InfoToTimestampCmd returns scp timestamp command for supplied info
func InfoToTimestampCmd(info os.FileInfo) string {
	unixTimestamp := info.ModTime().Unix()
	return fmt.Sprintf("T%v 0 %v 0\n", unixTimestamp, unixTimestamp)
}

//InfoToCreateCmd returns scp create command for supplied info
func InfoToCreateCmd(info os.FileInfo) string {
	mode := info.Mode()
	if mode >= 01000 { //symbolic linkg
		mode = DefaultFileMode
	}
	locationType := "C"
	size := info.Size()
	if info.IsDir() {
		locationType = "D"
		size = 0
	}
	fileMode := string(fmt.Sprintf("%v%04o", locationType, mode.Perm())[:5])
	return fmt.Sprintf("%v %d %s\n", fileMode, size, info.Name())
}
