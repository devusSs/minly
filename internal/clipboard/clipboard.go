package clipboard

import (
	"fmt"

	"github.com/atotto/clipboard"
)

func Write(text string) error {
	err := clipboard.WriteAll(text)
	if err != nil {
		return fmt.Errorf("failed to write to clipboard: %w", err)
	}

	return nil
}
