package shells

import (
	"fmt"
	"io"
)

// worker entry point to print stdout and stderr from remote programs.
func fabricPrinter(w io.Writer, printch <-chan string, kill chan bool) {
loop:
	for {
		select {
		case s, ok := <-printch:
			if !ok {
				break loop
			}
			fmt.Fprintf(w, "%v", s)
		case <-kill:
			break loop
		}
	}
}
