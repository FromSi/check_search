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

## Запросы для Meilisearch
```
# Найти по названию

curl \
  -X POST 'http://localhost:7700/indexes/idx_t1/search' \
  -H 'Content-Type: application/json' \
  --data-binary '{ "q": "dolor" }'
```
