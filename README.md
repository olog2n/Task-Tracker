# Sukno Tracker MVP

## Быстрый старт
```bash
# 1. Клонировать репо
git clone <repo>

# 2. Создать .env
cp .env.example .env
# Отредактировать JWT_SECRET

# 3. Собрать
docker-compose build

# 4. Запустить всё
docker-compose up -d

# 5. Открыть
# 👉 http://localhost:3000

## P.S.
# Где взять JWT_SECRET, обратите внимание на backend/scripts, там находится удобный скрипт для создания ключей, затем вам нужно будет лишь отредактировать .env