// Package resettest содержит структуры, покрывающие все правила генератора
// cmd/reset: примитивы, указатели, слайсы, мапы и вложенные структуры.
// Методы Reset() для них генерируются командой:
//
//	go run ./cmd/reset
package resettest

// Child — вложенная структура с генерируемым Reset().
//
// generate:reset
type Child struct {
	Value int
}

// Container покрывает все виды полей, поддерживаемые генератором.
//
// generate:reset
type Container struct {
	Num    int
	Str    string
	Flag   bool
	Ratio  float64
	StrPtr *string
	NumPtr *int
	Items  []string
	Index  map[string]int
	Child  Child
	ChildP *Child
}
