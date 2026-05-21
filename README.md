# Простой http сервер (net/http)

Реализованы операции GET, POST, PUT, DELETE над объектами `Note` (заметка), заметки хранятся в памяти пока запущен сервер (в map):

``` json
 // note
 { 
    "note_id": // уникальный id заметки (uuid)
    "header": // заголовок (not null)
    "body": // тело заметки
 }
```

А также запрос для конвертации валют, для получения актуального курса используется внешний api <https://freecurrencyapi.com/>

### Переменные окружения

```
# тип хранилища, может быть "memory" или "postgres"
STORAGE_TYPE=

# переменные для postgres должны совпадать с docker-compose
# при запуске сервера в контейнере, хост и порт заменяются в docker compose
POSTGRES_HOST=
POSTGRES_PORT=

POSTGRES_USER=
POSTGRES_PASSWORD=
POSTGRES_DB=
POSTGRES_SSLMODE=

# переменные для получения курсов валют
FREECURRENCY_API_URL=https://api.freecurrencyapi.com/v1/latest
FREECURRENCY_API_KEY=fca_live_8lWYrQ3cQDWZob9q0evmwYSrRYH6mtdPU5XTMfJc
```

---

### Запуск с линтером и тестами

#### В контейнере

``` bash
    make all
```

#### Без контейнера на windows (запускает postgres в отдельном контейнере)

``` bash
    make all-windows
```

---

## Эндпоинты

### Операции с заметками

- `GET /note` - все заметки, ответ - список Note

- `GET /note/header/{header}` - заметки с указанным заголовком, ответ - список Note

- `GET /note/id/{id}` - заметка по её id, ответ - объект Note

- `POST /note` - создание заметки, Note передается в теле, ответ - созданная заметка с id

- `PUT /note/{id}` - изменение заметки по id, Note передается в теле

- `DELETE /note/{id}` - удаление заметки по id

### Конвертация валют

`GET /currency?amount=&base=&currencies=`

Параметры (можно указать только нужные):

- `amount` - сумма конвертации, по умолчанию = 1
- `base` - исходная валюта, по умолчанию = RUB
- `currencies` список конечных валют (в виде `currencies="USD"&currencies="EUR"`), по умолчанию - все доступные

Пример ответа: \
`{"CNY":136.57400370000002,"EUR":17.06579925,"USD":19.9999971}`
