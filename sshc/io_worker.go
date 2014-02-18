package sshc

import (
	"bufio"
	"fmt"
	"github.com/couchbaselabs/cbsh/api"
	"io"
)

func readOut(rd io.Reader, ch outStr) {
	r := bufio.NewReader(rd)
	for {
		if buf, err := r.ReadBytes(api.NEWLINE); len(buf) > 0 {
			ch <- string(buf)
		} else if err != nil && err != io.EOF {
			ch <- fmt.Sprintf("%v", err)
			break
		} else {
			break
		}
	}
}

func writeIn(w io.WriteCloser, inch inStr, errch outStr) {
	for {
		if s, ok := <-inch; ok {
			bs := []byte(s)
			if _, err := w.Write(bs); err != nil {
				errch <- fmt.Sprintln(err)
			}
		} else {
			w.Close()
			break
		}
	}
}
