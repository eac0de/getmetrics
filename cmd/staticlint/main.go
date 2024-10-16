// Реализует запуск multichecker с набором анализаторов для статического анализа кода Go.
// В данном примере используются анализаторы из пакетов staticcheck, quickfix, а также стандартные анализаторы Go.
// Кроме того, подключен собственный анализатор noexit, который проверяет код на отсутствие вызовов os.Exit() в функциях.
// Для запуска multichecker необходимо в вызвать его с указанием пути до файла, который хотите проверить, для рекурсивной проверки из текущей дирекстории - `./...`
package main

import (
	"github.com/eac0de/getmetrics/cmd/staticlint/noexit"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/staticcheck"
)

// main является точкой входа программы.
// Функция main собирает набор анализаторов и запускает multichecker.
// Multichecker - это инструмент для запуска нескольких анализаторов на коде Go.
// Анализаторы позволяют находить различные ошибки, потенциальные баги и стилистические проблемы в коде.
func main() {

	var checks []*analysis.Analyzer

	// Добавление всех анализаторов из пакета staticcheck.
	// Staticcheck - это набор проверок, который находит баги, неэффективности,
	// нарушения стиля и другие потенциальные ошибки в Go-коде.
	for _, v := range staticcheck.Analyzers {
		checks = append(checks, v.Analyzer)
	}

	// Добавление всех анализаторов из пакета quickfix.
	// Quickfix предоставляет анализаторы, которые предлагают улучшения кода,
	// исправление потенциальных ошибок и повышение читаемости.
	for _, v := range quickfix.Analyzers {
		checks = append(checks, v.Analyzer)
	}

	// Добавление стандартных анализаторов из пакетов printf, shadow и structtag.

	// printf.Analyzer - проверяет правильность использования функции форматированного вывода (например, fmt.Printf).
	// Он проверяет корректность передаваемых аргументов для соответствующих форматирующих строк.
	checks = append(checks, printf.Analyzer)

	// shadow.Analyzer - обнаруживает переменные, которые затеняют (shadow) другие переменные в той же области видимости.
	// Это может привести к скрытым багам и путанице в коде.
	checks = append(checks, shadow.Analyzer)

	// structtag.Analyzer - проверяет правильность синтаксиса тегов структур.
	// Неправильные теги могут привести к неожиданным результатам при использовании таких библиотек, как encoding/json.
	checks = append(checks, structtag.Analyzer)

	// Добавление собственного анализатора noexit, который проверяет, что функции не вызывают os.Exit().
	// Это полезно в ситуациях, когда нужно гарантировать, что программа не завершится внезапно,
	// а будет корректно завершать выполнение через возврат ошибки или нормальный выход.
	checks = append(checks, noexit.Analyzer)

	multichecker.Main(checks...)
}
