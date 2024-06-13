## Установка
```
# для первого запуска
make install

# для последующих запусков
make run
```

```
t1=1_000pcs
t2=1_000_000pcs
```

## Ссылка на репозитории
* main - https://github.com/RediSearch/RediSearch
* doc - https://redis.io/docs/latest/develop/interact/search-and-query/
* rdi - https://redis.io/docs/latest/integrate/redis-data-integration/
* Redisearch-PHP - https://www.ethanhann.com/redisearch-php/
* main - https://github.com/meilisearch/meilisearch
* doc - https://www.meilisearch.com/docs
* Laravel Scout - https://laravel.com/docs/8.x/scout

## Версии полнотекстовых поисковиков
* Redisearch - 2.8.8 - Sep 27, 2023
* Meilisearch - 1.8.2 - Jun 10, 2024

## Первый релиз
* Redisearch - Nov 28, 2017 - 5k stars
* Meilisearch - Apr 3, 2023 - 44k stars

## Библиотеки для клиентов
* Redisearch - https://github.com/RediSearch/RediSearch/tree/v2.8.8?tab=readme-ov-file#client-libraries
* Meilisearch - https://www.meilisearch.com/docs/learn/what_is_meilisearch/sdks

## Схемы индексов для Redisearch
```
FT.CREATE idx_t1 
    ON HASH
    PREFIX 1 t1: 
SCHEMA 
    title TEXT SORTABLE
    description TEXT SORTABLE
    created_at NUMERIC SORTABLE
```

```
FT.CREATE idx_t2 
    ON HASH 
    PREFIX 1 t2: 
SCHEMA 
    t1_id NUMERIC SORTABLE
    title TEXT SORTABLE
    description TEXT SORTABLE
    data TEXT
    created_at NUMERIC SORTABLE
```

## Запросы для Redisearch
Для ввода нужно зайти через `make redisearch`

```
# Найти по названию

FT.SEARCH idx_t1 "dolor"
```

```
# Поиск по полю title

FT.SEARCH idx_t1 "@title:dolor"
```

```
# Поиск по полю title с NOT

FT.SEARCH idx_t1 "-nulla"
```

```
# Поиск по полю title с OR AND

FT.SEARCH idx_t1 "@title:dolor @description:(ex|et)"
```

```
# Поиск по полю created_at с BETWEEN

FT.SEARCH idx_t1 "@created_at:[652692981 938299797]"
```

```
# Поиск с пагинацией

FT.SEARCH idx_t1 "@created_at:[652692981 938299797]" SORTBY created_at LIMIT 0 5
```

```
# Поиск с агрегацией

FT.AGGREGATE idx_t1 "*" GROUPBY 1 @title REDUCE COUNT 0 AS num_records REDUCE MAX 1 @created_at AS max_created_at REDUCE MIN 1 @created_at AS min_created_at APPLY "timefmt(@max_created_at)" AS max_created_at APPLY "timefmt(@min_created_at)" AS min_created_at
```

## Запросы для Meilisearch
```
# Найти по названию

curl \
  -X POST 'http://localhost:7700/indexes/idx_t1/search' \
  -H 'Content-Type: application/json' \
  --data-binary '{ "q": "dolor" }'
```

```
# Поиск по полю title

curl \
  -X POST 'http://localhost:7700/indexes/idx_t1/search' \
  -H 'Content-Type: application/json' \
  --data-binary '{ "filter": "title = dolor" }'
```

```
# Поиск по полю title с NOT OR AND

curl \
  -X POST 'http://localhost:7700/indexes/idx_t1/search' \
  -H 'Content-Type: application/json' \
  --data-binary '{ "filter": "title != dolor AND (title = 'ex' OR title = 'et')" }'
```

```
# Поиск с гранями

curl \
  -X PUT 'http://localhost:7700/indexes/idx_t2/settings/filterable-attributes' \
  -H 'Content-Type: application/json' \
  --data-binary '[
    "title"
  ]'
  
curl \
  -X POST 'http://localhost:7700/indexes/idx_t2/search' \
  -H 'Content-Type: application/json' \
  --data-binary '{
        "q": "ex",
        "facets": [
            "title"
        ]
    }'
```

```
# Мульти-поиск

curl \
  -X POST 'http://localhost:7700/multi-search' \
  -H 'Content-Type: application/json' \
  --data-binary '{
    "queries": [
      {
        "indexUid": "idx_t1",
        "q": "et",
        "limit": 3
      },
      {
        "indexUid": "idx_t2",
        "q": "ex",
        "limit": 3
      }
    ]
  }'
```
