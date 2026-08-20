# Do Together
Do Together — backend API для совместного управления проектами. 
Он позволяет пользователям создавать проекты и задачи, добавлять участников, управлять их ролями и отслеживать ход работы.



## Возможности

- Регистрация и вход в систему.
- Создание и архивирование проектов.
- Получение доступных пользователю проектов с фильтрацией по статусу и пагинацией через `limit` и `offset`.
- Роли участников: `creator`, `admin`, `member`.
- Создание приглашений с указанием роли и срока действия.
- Создание, просмотр, редактирование и удаление задач, а также изменение их статуса.
- Обновление и завершение пользовательской сессии через refresh/logout.
- Защита HTTP-запросов с помощью access JWT.
- Хранение refresh-сессий в Redis с атомарной ротацией токенов.
- Управление участниками проекта и их ролями.
- Эндпоинты `/health` и `/ready` для проверки состояния приложения.

## Технологии

- Go — основной язык приложения.
- PostgreSQL — хранение постоянных данных.
- Redis — хранение refresh-сессий и атомарная ротация refresh-токенов.
- JWT — авторизация HTTP-запросов.
- REST API — организация взаимодействия клиента с приложением.
- golang-migrate — управление миграциями PostgreSQL.
- Docker и Docker Compose — сборка и локальный запуск приложения со всеми зависимостями.
- GitHub Actions — автоматическая проверка тестов, `go vet` и сборки приложения.

## Архитектура

Проект реализован как слоистый монолит. Код разделён на HTTP-обработчики, сервисы, доменные модели и репозитории.

```mermaid
flowchart LR
    Client["HTTP-клиент"] --> Router["Router и middleware"]
    Router --> Handler["Handler и DTO"]
    Handler --> Service["Service"]
    Service --> Domain["Domain"]
    Service --> Repository["Repository interface"]
    Repository --> PostgreSQL["PostgreSQL"]
    Repository --> Redis["Redis"]
    Service --> Handler
    Handler --> Client
```

Основной путь запроса:

`HTTP request → router/middleware → handler → request DTO → service → domain/repository → PostgreSQL или Redis → service → response DTO → HTTP response`


## Ключевые технические решения

- Принятие приглашения выполняется в транзакции: добавление участника и изменение статуса приглашения либо завершаются вместе, либо полностью откатываются.
- Права доступа дополнительно проверяются внутри SQL-запросов. Пользователь может работать только с проектами, в которых он состоит, с учётом своей роли.
- Refresh-сессии хранятся в Redis, а refresh-токены ротируются атомарно. При одновременном использовании одного токена успешно завершится только один запрос.
- При graceful shutdown сервер прекращает принимать новые запросы и предоставляет активным запросам ограниченное время для завершения.
- `/health` проверяет работу HTTP-приложения, а `/ready` дополнительно проверяет доступность PostgreSQL и Redis.



## Быстрый запуск

### Требования

- Git
- Docker Desktop с поддержкой Docker Compose

### 1. Клонирование репозитория

```bash
git clone https://github.com/glebH52-arch/petprojectjiro.git do-together
cd do-together
```

### 2. Настройка переменных окружения

Создайте `.env` на основе `.env.example`:

```bash
cp .env.example .env
```

На Windows PowerShell:

```powershell
Copy-Item .env.example .env
```

Укажите собственные значения как минимум для следующих переменных:

```env
JWT_SECRET=replace_with_secret_at_least_32_characters
POSTGRES_PASSWORD=replace_with_postgres_password
REDIS_PASSWORD=replace_with_redis_password
```

Файл `.env` содержит секреты и не должен добавляться в Git.

### 3. Запуск приложения

```bash
docker compose up -d --build
```

Docker Compose:

- запустит PostgreSQL и Redis;
- дождётся их готовности;
- применит миграции;
- соберёт и запустит HTTP-приложение.

Проверить состояние контейнеров:

```bash
docker compose ps -a
```

PostgreSQL, Redis и приложение должны иметь состояние `healthy`. Контейнер `migrate` должен завершиться со статусом `Exited (0)`.

### 4. Проверка приложения

```bash
curl http://localhost:8080/health
curl http://localhost:8080/ready
```

Ожидаемые ответы:

```json
{"status":"ready"}
```

```json
{"status":"ok"}
```

API доступен по адресу `http://localhost:8080`.

### 5. Остановка

```bash
docker compose down
```

Эта команда сохраняет данные PostgreSQL и Redis в Docker volumes.

## Переменные окружения

| Переменная | Назначение | Пример |
|---|---|---|
| `DATABASE_URL` | Строка подключения к PostgreSQL при локальном запуске без Compose | `postgres://postgres:change_me@localhost:5432/do_together?sslmode=disable` |
| `JWT_SECRET` | Секрет для подписи и проверки JWT, не менее 32 символов | `replace_with_secret_at_least_32_characters` |
| `ACCESS_TOKEN_TTL` | Время жизни access-токена | `15m` |
| `REDIS_ADDR` | Адрес Redis при локальном запуске | `localhost:6379` |
| `REDIS_PASSWORD` | Пароль для подключения к Redis | `change_me` |
| `REDIS_DB` | Номер логической базы Redis | `0` |
| `REFRESH_TOKEN_IDLE_TTL` | Срок неактивности refresh-сессии, обновляемый при успешной ротации | `168h` |
| `REFRESH_TOKEN_ABSOLUTE_TTL` | Максимальный срок жизни refresh-сессии с момента входа | `1440h` |
| `POSTGRES_USER` | Пользователь PostgreSQL для Docker Compose | `postgres` |
| `POSTGRES_PASSWORD` | Пароль пользователя PostgreSQL | `change_me` |
| `POSTGRES_DB` | Имя базы данных PostgreSQL | `do_together` |

При запуске через Docker Compose адреса PostgreSQL и Redis автоматически заменяются на адреса контейнеров `postgres:5432` и `redis:6379`.

## API

Все защищённые маршруты требуют заголовок:

```http
Authorization: Bearer <access_token>
```

### Авторизация и пользователи

| Метод | Маршрут | Назначение | JWT |
|---|---|---|---|
| `POST` | `/users` | Регистрация пользователя | Нет |
| `POST` | `/auth/login` | Вход и получение пары токенов | Нет |
| `POST` | `/auth/refresh` | Ротация refresh-токена и получение нового access-токена | Нет |
| `POST` | `/auth/logout` | Завершение refresh-сессии | Нет |
| `GET` | `/users/me` | Получение данных текущего пользователя | Да |

### Проекты

| Метод | Маршрут | Назначение | JWT |
|---|---|---|---|
| `POST` | `/projects` | Создание проекта | Да |
| `GET` | `/projects` | Список доступных проектов с фильтрацией и пагинацией | Да |
| `GET` | `/projects/{id}` | Получение проекта по ID | Да |
| `PATCH` | `/projects/{id}` | Редактирование проекта | Да |
| `DELETE` | `/projects/{id}` | Архивирование проекта | Да |

Параметры `GET /projects`:

- `status` — `active` или `archived`;
- `limit` — количество элементов;
- `offset` — смещение.

### Участники проекта

| Метод | Маршрут | Назначение | JWT |
|---|---|---|---|
| `GET` | `/projects/{id}/members` | Получение участников проекта | Да |
| `PATCH` | `/projects/{id}/members/{userID}` | Изменение роли участника | Да |
| `DELETE` | `/projects/{id}/members/{userID}` | Удаление участника | Да |

### Приглашения

| Метод | Маршрут | Назначение | JWT |
|---|---|---|---|
| `POST` | `/projects/{id}/invites` | Создание приглашения | Да |
| `GET` | `/invites` | Получение приглашений текущего пользователя | Да |
| `POST` | `/invites/{id}/accept` | Принятие приглашения | Да |
| `POST` | `/invites/{id}/decline` | Отклонение приглашения | Да |

### Задачи

| Метод | Маршрут | Назначение | JWT |
|---|---|---|---|
| `POST` | `/projects/{id}/tasks` | Создание задачи | Да |
| `GET` | `/projects/{id}/tasks` | Получение задач проекта | Да |
| `GET` | `/projects/{id}/tasks/{taskID}` | Получение задачи по ID | Да |
| `PATCH` | `/projects/{id}/tasks/{taskID}` | Редактирование задачи или её статуса | Да |
| `DELETE` | `/projects/{id}/tasks/{taskID}` | Удаление задачи | Да |

### Состояние приложения

| Метод | Маршрут | Назначение |
|---|---|---|
| `GET` | `/health` | Проверка работы HTTP-приложения |
| `GET` | `/ready` | Проверка готовности приложения, PostgreSQL и Redis |

## Примеры запросов

В примерах используется адрес `http://localhost:8080`.

### Регистрация

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "demo_user",
    "email": "demo@example.com",
    "password": "DemoPassword123!"
  }'
```

### Вход

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "demo@example.com",
    "password": "DemoPassword123!"
  }'
```

Ответ содержит access- и refresh-токены:

```json
{
  "access_token": "<access_token>",
  "refresh_token": "<refresh_token>",
  "token_type": "Bearer",
  "expires_in": 900
}
```

### Создание проекта

Замените `<access_token>` токеном, полученным при входе:

```bash
curl -X POST http://localhost:8080/projects \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Demo project",
    "goal": "Check the Do Together API"
  }'
```

### Получение проектов

```bash
curl "http://localhost:8080/projects?status=active&limit=20&offset=0" \
  -H "Authorization: Bearer <access_token>"
```

### Обновление токенов

Замените `<refresh_token>` актуальным refresh-токеном:

```bash
curl -X POST http://localhost:8080/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "<refresh_token>"
  }'
```

После успешной ротации старый refresh-токен больше не используется. Для следующего обновления необходимо сохранить новый токен из ответа.

### Выход

```bash
curl -X POST http://localhost:8080/auth/logout \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "<current_refresh_token>"
  }'
```

Успешный выход завершает текущую refresh-сессию и возвращает статус `204 No Content`.

## Проверки

В проекте настроен GitHub Actions workflow, который при pull request в `main` и после push в `main` выполняет:

```bash
go mod download
go test ./...
go vet ./...
go build -o do-together ./cmd
```

Автоматические тесты сейчас находятся в:

- `internal/domain/project_test.go`;
- `internal/repository/project_test.go`.

Остальная функциональность проверялась вручную через HTTP-запросы и Docker Compose, включая:

- регистрацию, вход, refresh и logout;
- создание проектов, приглашений и задач;
- изменение ролей и статусов;
- атомарную ротацию refresh-токенов;
- `/health` и `/ready`;
- сохранение PostgreSQL и Redis данных после перезапуска контейнеров.

## Ограничения

Текущая версия является MVP. В ней пока отсутствуют:

- подтверждение email;
- восстановление пароля;
- отправка приглашений по email — сейчас используется ID пользователя;
- rate limiting;
- frontend;
- деплой в публичное окружение;
- полное покрытие модулей unit- и интеграционными тестами.
