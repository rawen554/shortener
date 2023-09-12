# shortener

Сервис сокращения URL.

## Конфигурация сервиса

- адрес сервиса `flag:"a" env:"SERVER_ADDRESS"`
- базовый адрес для результирующей ссылки при сокращении`flag:"b" env:"BASE_URL"`
- путь в файловой системе для сохранения результатов в файл `flag:"f" env:"FILE_STORAGE_PATH"`
- адрес для подключения к БД `flag:"d" env:"DATABASE_DSN"`
- секрет, необходимый для создания jwt токенов `flag:"s" env:"SECRET"`

## Документация

Запустить `godoc -http:8080`
Перейти по адресу [http://localhost:8080/pkg/github.com/rawen554/shortener/?m=all](http://localhost:8080/pkg/github.com/rawen554/shortener/?m=all)

## Запуск сервиса

`go run cmd/staticlint/main.go`

## Сборка с версионированием
`go build -ldflags "-X main.buildVersion=0.0.1 -X 'main.buildDate=$(date +'%Y/%m/%d %H:%M:%S')' -X main.buildCommit=xxx" cmd/shortner/main.go`
