package scp

import (
	"github.com/viant/afs/file"
	"os"
	"strings"
	"time"
)

//adjustPath tracks current and previous relative path to adjust accordingly
func adjustPath(prev, current string, moveDown func(info os.FileInfo) error, moveUp func() error) error {
	if prev == current {
		return nil
	}
	var prevElements []string
	var currElements []string

	if prev != "" {
		prevElements = strings.Split(prev, "/")
	}
	if prev != "" {
		currElements = strings.Split(current, "/")
	}
	if len(prevElements) < len(currElements) {
		for i := len(prevElements); i < len(currElements); i++ {
			dirInfo := file.NewInfo(currElements[i], 0, file.DefaultDirOsMode, time.Now(), true)
			if err := moveDown(dirInfo); err != nil {
				return err
			}
		}
	}
	var downElements = make([]string, 0)
	for i := len(prevElements) - 1; i >= 0; i-- {
		prevElem := prevElements[i]
		currentElem := ""
		if i < len(currElements) {
			currentElem = currElements[i]
		}
		if currentElem == prevElem {
			break
		}
		if currentElem != "" {
			downElements = append(downElements, currentElem)
		}
		if err := moveUp(); err != nil {
			return err
		}
	}
	for _, element := range downElements {
		dirInfo := file.NewInfo(element, 0, file.DefaultDirOsMode, time.Now(), true)
		if err := moveDown(dirInfo); err != nil {
			return err
		}
	}
	return nil
}
