package extract

import "fmt"

/*
	Setting the IllegalPathError when an illegal
	path is detected during the extraction process.

	Typically, only the filename is shown on error,
	however, the absolute value of the illegal path
	may also be set.
*/

type IllegalPathError struct {
	Abs      string
	Filename string
}

func (e *IllegalPathError) Error() string {
	return fmt.Sprintf("Illegal path: %s", e.Abs)
}

func IsIllegalPathError(err error) bool {
	_, ok := err.(*IllegalPathError)
	return ok
}
