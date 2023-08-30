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
