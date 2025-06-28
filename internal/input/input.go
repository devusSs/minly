package input

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

// ReadFromStdin reads a line from standard input, prompting the user with the given message.
// It returns the input string with trimmed spaces or an error if reading fails.
func ReadFromStdin(prompt string) (string, error) {
	if prompt == "" {
		return "", errors.New("prompt cannot be empty")
	}

	if !isTerminal(os.Stdin) {
		return "", errors.New("stdin is not a terminal")
	}

	// Check if there is a space between the prompt and the input
	// and add one if there isn't.
	if !strings.HasSuffix(prompt, " ") {
		prompt += " "
	}

	fmt.Print(prompt)

	reader := bufio.NewReader(os.Stdin)
	str, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(str), nil
}

// isTerminal checks if the provided file descriptor is a terminal.
func isTerminal(f *os.File) bool {
	if f == nil {
		return false
	}

	info, err := f.Stat()
	if err != nil {
		return false
	}

	return info.Mode()&os.ModeCharDevice != 0
}
