package extract

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vbauerster/mpb/v7"
)

func TestExtract(t *testing.T) {
	testParent, err := ioutil.TempDir("", "extract_temp")
	if err != nil {
		t.Fatalf("Error creating temporary dir")
	}
	defer os.RemoveAll(testParent)
	for i, tc := range []struct {
		format   Format
		dest     string
		file     string
		expected bool
	}{
		{format: NewBz2(), dest: "Bz2", file: "testdata/test.bz2", expected: true},
		{format: NewGz(), dest: "Gz", file: "testdata/test.gz", expected: true},
		{format: NewRar(), dest: "Rar", file: "testdata/test.rar", expected: true},
		{format: NewTar(), dest: "Tar", file: "testdata/test.tar", expected: true},
		{format: NewTarGz(), dest: "Tgz", file: "testdata/test.tar.gz", expected: true},
		{format: NewZip(), dest: "Zip", file: "testdata/test.zip", expected: true},
	} {
		destDir := filepath.Join(testParent, tc.dest)
		start := time.Now()
		u, _ := tc.format.(Extractor)
		p := mpb.New()
		err := u.Extract(tc.file, destDir, p, start)
		if !tc.expected && err != nil {
			t.Errorf("[%d] [%s - %s] expected no error but got %s", i, tc.format, tc.file, err)
		}
	}
}

func TestMultipleTopLevels(t *testing.T) {
	for i, tc := range []struct {
		set    []string
		expect bool
	}{
		{
			set:    []string{},
			expect: false,
		},
		{
			set:    []string{"/a"},
			expect: false,
		},
		{
			set:    []string{"/a", "/a/b"},
			expect: false,
		},
		{
			set:    []string{"/a", "/b"},
			expect: true,
		},
		{
			set:    []string{"/a", "/ab"},
			expect: true,
		},
		{
			set:    []string{"a", "a/b"},
			expect: false,
		},
		{
			set:    []string{"a", "/a/b"},
			expect: false,
		},
		{
			set:    []string{"../a", "a/b"},
			expect: true,
		},
		{
			set:    []string{`C:\a\b`, `C:\a\b\c`},
			expect: false,
		},
		{
			set:    []string{`C:\`, `C:\a\b`},
			expect: false,
		},
		{
			set:    []string{`D:\a`, `E:\a`},
			expect: true,
		},
		{
			set:    []string{`D:\a`, `E:\a`, `C:\a`},
			expect: true,
		},
		{
			set:    []string{"/a", "/", "/b"},
			expect: true,
		},
	} {
		actual := TopLevels(tc.set)
		if actual != tc.expect {
			t.Errorf("Test %d: %v: Expected %t, got %v", i, tc.set, tc.expect, actual)
		}
	}
}
