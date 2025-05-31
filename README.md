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
  - cn (text)

- usersmap
  - id
  - uid

- passwords
  - id
  - login (blob)
  - password (blob)
  - meta (blob)

- passwordsmap

- text
  - id
  - text (blob)
  - meta (blob)

- textmap

- bin
  - id
  - data (blob)
  - meta (blob)

- binmap

- bank
  - id
  - number
  - date
  - cvv
  - meta

- bankmap

## Клиент

register - генерация ключа, запроса на сертификат, получение клиентского и CA сертификатов, сохранение их в бандл

interactive - список записей, выбор и просмотр/редактирование/удаление записи, добавление записи
