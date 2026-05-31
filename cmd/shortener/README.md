# cmd/shortener

В данной директории содержится код, который скомпилируется в бинарное приложение.

Рекомендуется помещать только код, необходимый для запуска приложения, но не бизнес-логику.

Название директории должно соответствовать названию приложения.

Директория `cmd/shortener` содержит:
- точку входа в приложение (функция `main`)
- инициализацию зависимостей (можно вынести в отдельный пакет `internal/app`)
- настройку и запуск HTTP-сервера (можно вынести в отдельный пакет `internal/router`)
- обработку сигналов завершения работы приложения

## Оптимизация использования памяти (Профилирование pprof)

В рамках оптимизации сервиса было проведено профилирование памяти (alloc_space) с использованием встроенного инструмента `pprof`. 

### Выявленные проблемы:
1. Метод `FileRepository.Save` открывал и закрывал файл на диске при каждом вызове, что создавало избыточные системные аллокации.
2. Метод `GetURLsByUserID` производил полное чтение и парсинг файла с диска вместо использования кэша в оперативной памяти.

### Результаты оптимизации:
После вынесения дескриптора файла и JSON-энкодера в структуру репозитория (однократная инициализация), а также перевода чтения пользователя на RAM-кэш, повторный замер бенчмарков показал значительное снижение потребления памяти:

      flat  flat%   sum%        cum   cum%
-6656.51kB  1.19% 43.27% -6656.51kB  1.19%  syscall.ByteSliceFromString
-3072.24kB  0.55% 49.22% -3072.24kB  0.55%  os.newFile
-2560.12kB  0.46% 49.32% -2560.12kB  0.46%  github.com/google/uuid.UUID.String (inline)
-2048.12kB  0.37% 49.86% -2560.25kB  0.46%  bytes.(*Buffer).grow
-1970.09kB  0.35% 50.98% -11697.96kB  2.09%  github.com/sleepy-moon-cake/golang_transform_url/internal/repository.(*FileRepository).Save
-1536.03kB  0.27% 51.53% 51291.02kB  9.17%  github.com/sleepy-moon-cake/golang_transform_url/internal/shared/audit.NewAuditMiddleware.func1.1
-1536.02kB  0.27% 51.25% -1536.02kB  0.27%  io.NopCloser (inline)
         0     0% 51.43%  -2048.12kB  0.37%  bytes.(*Buffer).Write
         0     0% 51.43%  -1536.10kB  0.27%  github.com/google/uuid.NewString
         0     0% 51.43%  -2048.12kB  0.37%  github.com/sleepy-moon-cake/golang_transform_url/internal/logger.(*loggerResponse).Write
         0     0% 51.43%  -2048.12kB  0.37%  github.com/sleepy-moon-cake/golang_transform_url/internal/shared/audit.(*ResponseWrapper).Write
         0     0% 51.43%  -2048.12kB  0.37%  net/http/httptest.(*ResponseRecorder).Write
         0     0% 51.43%  -9728.75kB  1.74%  os.OpenFile
         0     0% 51.43%  -6656.51kB  1.19%  os.ignoringEINTR (inline)
         0     0% 51.43%  -6656.51kB  1.19%  os.open (inline)
         0     0% 51.43%  -9728.75kB  1.74%  os.openFileNolog
         0     0% 51.43%  -6656.51kB  1.19%  os.openFileNolog.func1 (inline)
         0     0% 51.43%  -6656.51kB  1.19%  syscall.BytePtrFromString (inline)
         0     0% 51.43%  -6656.51kB  1.19%  syscall.Open
