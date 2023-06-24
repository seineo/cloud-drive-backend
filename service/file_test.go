package service

import (
	"os"
	"path/filepath"
	"testing"
)

func TestArchiveFile(t *testing.T) {
	type args struct {
		location string
		fileName string
		dstPath  string
	}
	// set test directory
	testDir := "./test-archive-file"
	err := os.Mkdir(testDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test_normal",
			args: args{
				location: filepath.Join(testDir, "fileHash1"),
				fileName: "test_normal",
				dstPath:  filepath.Join(testDir, "test_normal.zip"),
			},
			wantErr: false,
		},
		{
			name: "test_file_not_exists",
			args: args{
				location: filepath.Join(testDir, "fileHash2"),
				fileName: "test_file_not_exists",
				dstPath:  filepath.Join(testDir, "test_file_not_exists.zip"),
			},
			wantErr: true,
		},
	}
	// create test files and write its filename as content
	for _, test := range tests {
		if !test.wantErr {
			file, err := os.Create(test.args.location)
			if err != nil {
				t.Fatal(err)
			}
			_, err = file.WriteString(test.args.fileName)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	// run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ArchiveFile(tt.args.location, tt.args.fileName, tt.args.dstPath); (err != nil) != tt.wantErr {
				t.Errorf("ArchiveFile() error = %v, wantErr %v", err, tt.wantErr)
			} else { // check whether zip file exists
				_, err := os.Stat(tt.args.dstPath)
				if err != nil {
					t.Errorf("ArchiveFile() err = %v", err)
				}
			}
		})
	}
	// delete test files
	err = os.RemoveAll(testDir)
	if err != nil {
		t.Fatal(err)
	}
}
