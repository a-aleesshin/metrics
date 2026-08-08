package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("start")

	if len(os.Args) > 1 {
		os.Exit(1) // want `прямой вызов os.Exit в функции main пакета main запрещён`
	}

	defer func() {
		os.Exit(2) // want `прямой вызов os.Exit в функции main пакета main запрещён`
	}()

	helper()
}

// helper — не main, вызов здесь разрешён.
func helper() {
	os.Exit(3)
}
