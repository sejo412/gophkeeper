# gophkeeper

Шифрование/дешифрование на стороне клиента

Аутентификаци/авторизация по сертификатам

## Сервер

GRPC:

- TLS CRUD для данных пользователя

- No-TLS регистрация пользователя (получение CA и клиентского сертификата)

СУБД:

- users
  - id (int)
  - uid (int)
  - cn (text)

- passwords
  - id
  - uid (int)
  - login (blob)
  - password (blob)
  - meta (blob)

- text
  - id
  - uid (int)
  - text (blob)
  - meta (blob)

- bin
  - id
  - uid (int)
  - data (blob)
  - meta (blob)

- bank
  - id
  - uid (int)
  - number
  - date
  - cvv
  - meta

## Клиент

register - генерация ключа, запроса на сертификат, получение клиентского и CA сертификатов, сохранение их в бандл

interactive - список записей, выбор и просмотр/редактирование/удаление записи, добавление записи
