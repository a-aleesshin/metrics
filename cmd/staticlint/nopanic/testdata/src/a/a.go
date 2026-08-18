package a

import "errors"

func direct() {
	panic("boom") // want `обнаружен вызов встроенной функции panic`
}

func withValue() {
	panic(errors.New("boom")) // want `обнаружен вызов встроенной функции panic`
}

func inGoroutine() {
	go func() {
		panic("boom") // want `обнаружен вызов встроенной функции panic`
	}()
}

func inDefer() {
	defer panic("boom") // want `обнаружен вызов встроенной функции panic`
}

// shadowed — одноимённая пользовательская функция не должна репортиться.
func shadowed() {
	panic := func(string) {}
	panic("not a builtin")
}

func clean() error {
	return errors.New("ошибки возвращаем значениями")
}
