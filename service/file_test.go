package service

import (
	"testing"
)

func TestArchiveFile(t *testing.T) {
	root := "/Users/liyuewei/Desktop/test-zip"
	target := "/Users/liyuewei/Desktop/test-zip-result.zip"
	err := ArchiveFile(root, target)
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log("successfully archived files")
	}
}
