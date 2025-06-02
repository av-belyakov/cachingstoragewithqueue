package supportingfunctions

import (
	"fmt"
	"runtime"
)

// CustomError ошибка дополняется ссылкой на файл и номером строки в файле
func CustomError(err error) error {
	_, f, l, _ := runtime.Caller(1)

	return fmt.Errorf("%w %s:%d", err, f, l)
}
