# Task Service

Сервис для управления задачами с HTTP API на Go.

## Требования

- Go `1.23+`
- Docker и Docker Compose

## Быстрый запуск через Docker Compose

```bash
docker compose up --build
```

После запуска сервис будет доступен по адресу `http://localhost:8080`.

Если `postgres` уже запускался ранее со старой схемой, пересоздай volume:

```bash
docker compose down -v
docker compose up --build
```

Причина в том, что SQL-файл из `migrations/0001_create_tasks.up.sql` монтируется в `docker-entrypoint-initdb.d` и применяется только при инициализации пустого data volume.

## Swagger

Swagger UI:

```text
http://localhost:8080/swagger/
```

OpenAPI JSON:

```text
http://localhost:8080/swagger/openapi.json
```

## API

Базовый префикс API:

```text
/api/v1
```

Основные маршруты:

- `POST /api/v1/tasks`
- `GET /api/v1/tasks`
- `GET /api/v1/tasks/{id}`
- `PUT /api/v1/tasks/{id}`
- `DELETE /api/v1/tasks/{id}`


# Task Management Service (Extended)

Тестовое задание на позицию Junior Backend Developer (Go).  
В проект добавлена функциональность для работы с **периодическими задачами**.

## Что было реализовано

### 1. Схема базы данных (Database Schema)
- Создана новая таблица `task_recurrences` для хранения правил повторения задач.
- Реализована связь **One-to-One** с основной таблицей `tasks`.
- Настроено каскадное удаление (`ON DELETE CASCADE`): при удалении основной задачи правила её повторения удаляются автоматически.

### 2. Слой Domain (Модели)
- В структуру `Task` добавлено вложенное поле `Recurrence`.
- Создана структура `TaskRecurrence`, описывающая тип повторения (`daily`, `monthly`, `parity` и др.), интервал и дополнительные значения.
- Использованы теги `json:"...,omitempty"`, чтобы не загромождать ответ API, если у задачи нет повторений.

### 3. Слой Repository (Infrastructure)
- Обновлен метод `Create`: теперь сохранение задачи и её периодичности происходит в рамках работы с БД (используется `RETURNING id` для связки таблиц).
- Обновлен метод `GetByID` и `List`: реализовано получение данных через **LEFT JOIN**, что позволяет достать задачу и её настройки за один SQL-запрос.

### 4. API & Transport
- Настроена маршрутизация для работы через префикс `/api/v1/`.
- Поддержан прием вложенных JSON-объектов в теле POST-запроса.

## Как запустить

1. Убедитесь, что у вас установлен Docker и Docker Compose.
2. Запустите проект командой:
   ```bash
   docker-compose up --build -d