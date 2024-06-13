package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-faker/faker/v4"
	"github.com/go-redis/redis/v8"
	"github.com/meilisearch/meilisearch-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

type T1 struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Title       string `gorm:"type:varchar(255);not null" json:"title"`
	Description string `json:"description"`
	CreatedAt   string `gorm:"type:timestamp;not null" json:"created_at"`
	T2s         []T2   `gorm:"foreignKey:T1ID" json:"t2s"`
}

type T2 struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	T1ID        uint   `gorm:"not null" json:"t1_id"`
	Title       string `gorm:"type:varchar(255);not null" json:"title"`
	Description string `json:"description"`
	Data        string `gorm:"type:jsonb;not null" json:"data"`
	CreatedAt   string `gorm:"type:timestamp;not null" json:"created_at"`
	T1          T1     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:T1ID;references:ID" json:"-"`
}

func generateRandomData() string {
	rand.Seed(time.Now().UnixNano())
	numKeys := rand.Intn(10) + 1
	data := make(map[string]interface{})

	for i := 0; i < numKeys; i++ {
		key := strconv.Itoa(rand.Intn(100) + 1)

		if rand.Intn(2) == 0 {
			data[key] = faker.Word()
		} else {
			numValues := rand.Intn(11)
			values := make([]int, numValues)

			for j := 0; j < numValues; j++ {
				values[j] = rand.Intn(100) + 1
			}

			data[key] = values
		}
	}

	jsonData, err := json.Marshal(data)

	if err != nil {
		log.Fatalf("failed to marshal data: %v", err)
	}

	return string(jsonData)
}

func main() {
	time.Sleep(3 * time.Second)

	ctx := context.Background()

	// Подключение к PostgreSQL
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to PostgreSQL: %v", err)
	}

	db.AutoMigrate(&T1{}, &T2{})

	// Подключение к Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
	})

	// Проверка соединения с Redis
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}

	err = rdb.Do(ctx, "FT.CREATE", "idx_t1", "ON", "HASH", "PREFIX", "1", "t1:", "SCHEMA", "title", "TEXT", "SORTABLE", "description", "TEXT", "SORTABLE", "created_at", "NUMERIC", "SORTABLE").Err()
	if err != nil {
		log.Fatalf("failed create index to Redis: %v", err)
	}

	err = rdb.Do(ctx, "FT.CREATE", "idx_t2", "ON", "HASH", "PREFIX", "1", "t2:", "SCHEMA", "t1_id", "NUMERIC", "SORTABLE", "title", "TEXT", "SORTABLE", "description", "TEXT", "SORTABLE", "data", "TEXT", "created_at", "NUMERIC", "SORTABLE").Err()
	if err != nil {
		log.Fatalf("failed create index to Redis: %v", err)
	}

	// Подключение к Meilisearch
	ms := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   fmt.Sprintf("http://%s:%s", os.Getenv("MEILI_HOST"), os.Getenv("MEILI_PORT")),
		APIKey: "",
	})

	// Создание нового индекса в Meilisearch
	msIdxT1 := ms.Index("idx_t1")
	msIdxT2 := ms.Index("idx_t2")

	var wg sync.WaitGroup

	// Функция для записи данных в PostgreSQL, Redis и Meilisearch
	saveData := func(t1 T1, t2s []T2) {
		// Начало транзакции
		tx := db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				log.Fatalf("transaction failed: %v", r)
			}
		}()

		// Сохранение данных в PostgreSQL
		if err := tx.Model(&T1{}).Create(&t1).Error; err != nil {
			tx.Rollback()
			log.Fatalf("failed to create t1 record: %v", err)
		}
		for index, t2 := range t2s {
			t2.T1ID = t1.ID
			if err := tx.Model(&T2{}).Create(&t2).Error; err != nil {
				tx.Rollback()
				log.Fatalf("failed to create t2 record: %v", err)
			}
			t2s[index].ID = t2.ID
		}

		// Коммит транзакции PostgreSQL
		if err := tx.Commit().Error; err != nil {
			log.Fatalf("failed to commit transaction: %v", err)
		}

		tl := "2006-01-02 15:04:05"

		// Запись данных в Redis
		t1Key := fmt.Sprintf("t1:%d", t1.ID)
		t, _ := time.Parse(tl, t1.CreatedAt)
		t1Data := []interface{}{
			"title", t1.Title,
			"description", t1.Description,
			"created_at", t.Unix(),
		}
		pipe := rdb.TxPipeline()
		pipe.HSet(ctx, t1Key, t1Data...)
		for _, t2 := range t2s {
			t2Key := fmt.Sprintf("t2:%d", t2.ID)
			t, _ := time.Parse(tl, t2.CreatedAt)
			t2Data := []interface{}{
				"t1_id", strconv.Itoa(int(t2.T1ID)),
				"title", t2.Title,
				"description", t2.Description,
				"data", t2.Data,
				"created_at", t.Unix(),
			}
			pipe.HSet(ctx, t2Key, t2Data...)
		}
		_, err = pipe.Exec(ctx)
		if err != nil {
			log.Fatalf("failed to execute Redis pipeline: %v", err)
		}

		// Запись данных в Meilisearch
		if _, err := msIdxT1.AddDocuments([]T1{t1}); err != nil {
			log.Fatalf("Failed t1 to add documents to Meilisearch: %v", err)
		}
		if _, err := msIdxT2.AddDocuments(t2s); err != nil {
			log.Fatalf("Failed t2s to add documents to Meilisearch: %v", err)
		}
	}

	// Генерация и сохранение данных
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 1; i <= 20; i++ {
				t1 := T1{
					Title:       faker.Word(),
					Description: faker.Paragraph(),
					CreatedAt:   faker.Timestamp(),
				}

				var t2s []T2
				for j := 1000; j <= 1; j++ {
					t2 := T2{
						Title:       faker.Word(),
						Description: faker.Paragraph(),
						Data:        generateRandomData(),
						CreatedAt:   faker.Timestamp(),
					}
					t2s = append(t2s, t2)
				}

				saveData(t1, t2s)
			}
		}()
	}

	wg.Wait()
	fmt.Println("Data has been populated successfully.")
}
