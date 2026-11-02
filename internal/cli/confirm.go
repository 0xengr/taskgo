package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

// confirm writes a prompt to out and reads one answer from in. End of input is
// treated as "no", never as an implicit yes.
func confirm(in io.Reader, out io.Writer, prompt string) (bool, error) {
	if _, err := fmt.Fprint(out, prompt); err != nil {
		return false, err
	}

	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}

	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}
