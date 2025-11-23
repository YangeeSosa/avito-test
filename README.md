# PR Reviewer Assignment Service

Решение тестового задания Avito. Реализовано как in-memory сервис на Go.

## Запуск

```bash
make run
```

или через Docker (порт 8080):

```bash
docker-compose up --build
```

## Доступные ручки

| Метод | URL | Назначение |
|-------|-----|------------|
| POST | /team/add | Создать команду с участниками |
| GET | /team/get?team_name= | Получить команду и участников |
| POST | /users/setIsActive | Включить/выключить пользователя |
| POST | /pullRequest/create | Создать PR, автоматически назначить ревьюверов |
| POST | /pullRequest/merge | Перевести PR в MERGED (идемпотентно) |
| POST | /pullRequest/reassign | Переназначить ревьювера в рамках его команды |
| GET | /users/getReview?user_id= | Список PR, где пользователь ревьювер |
| GET | /health | health-check |

## Устройство

- `internal/models` - доменные сущности.
- `internal/db` - in-memory хранилище с мьютексом и индексом по ревьюверам.
- `internal/service` - бизнес-правила назначения/переназначения и валидация доменных ограничений.
- `internal/http` - chi-роутер, сериализация/десериализация DTO и маппинг ошибок в коды OpenAPI.
- `cmd/server` - точка входа.
