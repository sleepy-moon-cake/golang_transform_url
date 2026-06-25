// Пакет main представляет собой кастомный инструмент статического анализа (multichecker),
// объединяющий в себе стандартные линтеры Go, правила staticcheck и сторонние анализаторы.
//
// # Механизм запуска
//
// Для компиляции утилиты выполните команду из корня проекта:
//
//	go build -o ./cmd/staticlint/staticlint ./cmd/staticlint/main.go
//
// После сборки вы можете запустить анализ всего проекта или конкретных пакетов:
//
//	./cmd/staticlint/staticlint ./...
//
// # Состав анализаторов
//
// 1. Стандартные анализаторы пакета golang.org/x/tools/go/analysis/passes:
//   - asmdecl: Проверяет соответствие между объявлениями ASM и функциями Go.
//   - assign: Обнаруживает бесполезные присваивания (например, x = x).
//   - atomic: Проверяет корректность использования пакета sync/atomic.
//   - bools: Находит ошибки в логических выражениях.
//   - buildtag: Проверяет правильность написания build tags.
//   - cgocall: Проверяет правила передачи указателей в cgo.
//   - composite: Проверяет неинициализированные составные литералы.
//   - copylocks: Обнаруживает копирование блокировок (Mutex) по значению.
//   - errorsas: Проверяет, что второй аргумент errors.As является указателем на ошибку.
//   - httpresponse: Проверяет, что Response.Body закрывается вовремя.
//   - loopclosure: Проверяет ссылки на переменные цикла внутри горутин/отложенных вызовов.
//   - lostcancel: Проверяет вызов функции cancel() для контекстов.
//   - nilfunc: Находит бесполезные сравнения функций с nil.
//   - printf: Проверяет соответствие строк форматирования Printf их аргументам.
//   - shadow: Находит затененные (shadowed) переменные.
//   - shift: Проверяет сдвиги, превышающие размер типа.
//   - stdmethods: Проверяет сигнатуры стандартных методов (например, Read, Write).
//   - stringintconv: Находит сомнительные преобразования чисел в строки.
//   - structtag: Проверяет синтаксис тегов структур.
//   - tests: Проверяет правильность именования и сигнатур тестов.
//   - unmarshal: Проверяет корректность передачи типов в json.Unmarshal.
//   - unreachable: Находит недостижимый код.
//   - unsafeptr: Проверяет недопустимые преобразования uintptr в unsafe.Pointer.
//   - unusedresult: Находит неиспользуемые результаты вызовов функций.
//
// 2. Анализаторы пакета staticcheck.io:
//   - Все анализаторы класса SA (Static Analysis) — поиск явных багов и логических ошибок.
//   - Анализатор S1000 (класс S) — использование канала вместо select с одним кейсом.
//   - Анализатор ST1000 (класс ST) — проверка наличия документации пакета.
//   - Анализатор QF1001 (класс QF) — предложение использовать strings.Cut вместо strings.Index.
//
// 3. Публичные сторонние анализаторы:
//   - bodyclose (://github.com): Проверяет, закрыты ли тела ответов HTTP (res.Body.Close()).
//   - errwrap (://github.com): Находит места, где ошибки можно обернуть через %w вместо %v.
//
// 4. Собственный анализатор:
//   - osexit: Запрещает совершать прямой вызов os.Exit внутри функции main пакета main.
package main

import (
	"github.com/fatih/errwrap/errwrap"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func main() {
	var myAnalyzers []*analysis.Analyzer

	// 1. Добавляем стандартные анализаторы из passes
	myAnalyzers = append(myAnalyzers,
		asmdecl.Analyzer, assign.Analyzer, atomic.Analyzer, bools.Analyzer,
		buildtag.Analyzer, cgocall.Analyzer, composite.Analyzer, copylock.Analyzer,
		errorsas.Analyzer, httpresponse.Analyzer, loopclosure.Analyzer, lostcancel.Analyzer,
		nilfunc.Analyzer, printf.Analyzer, shadow.Analyzer, shift.Analyzer,
		stdmethods.Analyzer, stringintconv.Analyzer, structtag.Analyzer, tests.Analyzer,
		unmarshal.Analyzer, unreachable.Analyzer, unsafeptr.Analyzer, unusedresult.Analyzer,
	)

	// 2. Добавляем ВСЕ анализаторы класса SA из staticcheck
	for _, v := range staticcheck.Analyzers {
		myAnalyzers = append(myAnalyzers, v.Analyzer)
	}

	// 3. Добавляем по одному анализатору из других классов staticcheck
	// Класс S (Simple)
	for _, v := range simple.Analyzers {
		if v.Analyzer.Name == "S1000" {
			myAnalyzers = append(myAnalyzers, v.Analyzer)
			break
		}
	}
	// Класс ST (Style)
	for _, v := range stylecheck.Analyzers {
		if v.Analyzer.Name == "ST1000" {
			myAnalyzers = append(myAnalyzers, v.Analyzer)
			break
		}
	}
	// Класс QF (Quickfix)
	for _, v := range quickfix.Analyzers {
		if v.Analyzer.Name == "QF1001" {
			myAnalyzers = append(myAnalyzers, v.Analyzer)
			break
		}
	}

	// 4. Добавляем сторонние публичные анализаторы
	myAnalyzers = append(myAnalyzers, bodyclose.Analyzer)
	myAnalyzers = append(myAnalyzers, errwrap.Analyzer)

	// 5. Добавляем собственный анализатор os.Exit
	myAnalyzers = append(myAnalyzers, OsExitAnalizer)

	// Запуск мультичекера
	multichecker.Main(myAnalyzers...)
}
