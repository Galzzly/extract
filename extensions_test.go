package extract

import "testing"

func TestFileFormat(t *testing.T) {
	for i, tc := range []struct {
		checker   Format
		file      string
		shouldErr bool
	}{
		{checker: NewBz2(), file: "testdata/test.bz2", shouldErr: false},
		{checker: NewBz2(), file: "testdata/test.gz", shouldErr: true},
		{checker: NewBz2(), file: "testdata/test.rar", shouldErr: true},
		{checker: NewBz2(), file: "testdata/test.tar", shouldErr: true},
		{checker: NewBz2(), file: "testdata/test.tar.gz", shouldErr: true},
		{checker: NewBz2(), file: "testdata/test.tgz", shouldErr: true},
		{checker: NewBz2(), file: "testdata/test.zip", shouldErr: true},
		{checker: NewBz2(), file: "testdata/test.txt", shouldErr: true},

		{checker: NewGz(), file: "testdata/test.bz2", shouldErr: true},
		{checker: NewGz(), file: "testdata/test.gz", shouldErr: false},
		{checker: NewGz(), file: "testdata/test.rar", shouldErr: true},
		{checker: NewGz(), file: "testdata/test.tar", shouldErr: true},
		{checker: NewGz(), file: "testdata/test.tar.gz", shouldErr: false},
		{checker: NewGz(), file: "testdata/test.tgz", shouldErr: false},
		{checker: NewGz(), file: "testdata/test.zip", shouldErr: true},
		{checker: NewGz(), file: "testdata/test.txt", shouldErr: true},

		{checker: NewRar(), file: "testdata/test.bz2", shouldErr: true},
		{checker: NewRar(), file: "testdata/test.gz", shouldErr: true},
		{checker: NewRar(), file: "testdata/test.rar", shouldErr: false},
		{checker: NewRar(), file: "testdata/test.tar", shouldErr: true},
		{checker: NewRar(), file: "testdata/test.tar.gz", shouldErr: true},
		{checker: NewRar(), file: "testdata/test.tgz", shouldErr: true},
		{checker: NewRar(), file: "testdata/test.zip", shouldErr: true},
		{checker: NewRar(), file: "testdata/test.txt", shouldErr: true},

		{checker: NewTar(), file: "testdata/test.bz2", shouldErr: true},
		{checker: NewTar(), file: "testdata/test.gz", shouldErr: true},
		{checker: NewTar(), file: "testdata/test.rar", shouldErr: true},
		{checker: NewTar(), file: "testdata/test.tar", shouldErr: false},
		{checker: NewTar(), file: "testdata/test.tar.gz", shouldErr: true},
		{checker: NewTar(), file: "testdata/test.tgz", shouldErr: true},
		{checker: NewTar(), file: "testdata/test.zip", shouldErr: true},
		{checker: NewTar(), file: "testdata/test.txt", shouldErr: true},

		{checker: NewTarGz(), file: "testdata/test.bz2", shouldErr: true},
		{checker: NewTarGz(), file: "testdata/test.gz", shouldErr: true},
		{checker: NewTarGz(), file: "testdata/test.rar", shouldErr: true},
		{checker: NewTarGz(), file: "testdata/test.tar", shouldErr: true},
		{checker: NewTarGz(), file: "testdata/test.tar.gz", shouldErr: false},
		{checker: NewTarGz(), file: "testdata/test.tgz", shouldErr: false},
		{checker: NewTarGz(), file: "testdata/test.zip", shouldErr: true},
		{checker: NewTarGz(), file: "testdata/test.txt", shouldErr: true},

		{checker: NewZip(), file: "testdata/test.bz2", shouldErr: true},
		{checker: NewZip(), file: "testdata/test.gz", shouldErr: true},
		{checker: NewZip(), file: "testdata/test.rar", shouldErr: true},
		{checker: NewZip(), file: "testdata/test.tar", shouldErr: true},
		{checker: NewZip(), file: "testdata/test.tar.gz", shouldErr: true},
		{checker: NewZip(), file: "testdata/test.tgz", shouldErr: true},
		{checker: NewZip(), file: "testdata/test.zip", shouldErr: false},
		{checker: NewZip(), file: "testdata/test.txt", shouldErr: true},
	} {
		err := tc.checker.CheckFormat(tc.file)
		if tc.shouldErr && err == nil {
			t.Errorf("[%d] [%s - %s] expected error but got nil", i, tc.checker, tc.file)
		}
		if !tc.shouldErr && err != nil {
			t.Errorf("[%d] [%s - %s] expected no error but got %s", i, tc.checker, tc.file, err)
		}
	}
}
