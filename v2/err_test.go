package extract

import (
	"errors"
	"fmt"
	"os"
	"testing"
)

func TestIllegalPathErrorString(t *testing.T) {
	tests := []struct {
		instance *IllegalPathError
		expected string
	}{
		{instance: &IllegalPathError{Filename: "foo.txt"}, expected: "Illegal path: foo.txt"},
		{instance: &IllegalPathError{Abs: "/tmp/bar.txt", Filename: "bar.txt"}, expected: "Illegal path: bar.txt"},
	}

	for i, test := range tests {
		test := test
		t.Run(fmt.Sprintf("Case %d", i), func(t *testing.T) {
			if test.expected != test.instance.Error() {
				t.Fatalf("Expected '%s', but got '%s'", test.expected, test.instance.Error())
			}
		})
	}
}

func TestIsIllegalPathError(t *testing.T) {
	tests := []struct {
		instance error
		expected bool
	}{
		{instance: nil, expected: false},
		{instance: os.ErrNotExist, expected: false},
		{instance: fmt.Errorf("some error"), expected: false},
		{instance: errors.New("another error"), expected: false},
		{instance: &IllegalPathError{Filename: "foo.txt"}, expected: true},
	}

	for i, test := range tests {
		test := test

		t.Run(fmt.Sprintf("Case %d", i), func(t *testing.T) {
			actual := IsIllegalPathError(test.instance)
			if actual != test.expected {
				t.Fatalf("Expected '%v', but got '%v'", test.expected, actual)
			}
		})
	}
}
