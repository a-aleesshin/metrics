// Package b — не main: os.Exit здесь разрешён даже в функции main.
package b

import "os"

func main() {
	os.Exit(1)
}
