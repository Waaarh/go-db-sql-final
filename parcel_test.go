package main

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
	//"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

func SetupTestDB(t *testing.T) (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`
CREATE TABLE parcel (
    number INTEGER PRIMARY KEY AUTOINCREMENT,
    client INTEGER NOT NULL,
    status TEXT NOT NULL,
    address TEXT NOT NULL,
    created_at TEXT NOT NULL
	);
`)
	if err != nil {
		t.Fatal(err)
	}
	return db, err
}

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	db, err := SetupTestDB(t)
	assert.NoError(t, err)
	defer db.Close()
	store := ParcelStore{db}
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	assert.NoError(t, err)
	assert.Greater(t, id, 0)
	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	cash, err := store.Get(id)
	assert.NoError(t, err)

	assert.Equal(t, id, cash.Number)
	assert.Equal(t, parcel.Client, cash.Client)
	assert.Equal(t, parcel.Status, cash.Status)
	assert.Equal(t, parcel.Address, cash.Address)
	assert.Equal(t, parcel.CreatedAt, cash.CreatedAt)

	err = store.Delete(id)
	assert.NoError(t, err)

	_, err = store.Get(id)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := SetupTestDB(t)
	assert.NoError(t, err)
	defer db.Close()
	store := ParcelStore{db}
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	assert.NoError(t, err)
	assert.Greater(t, id, 0)
	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	assert.NoError(t, err)

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	Wow, err := store.Get(id)
	assert.NoError(t, err)
	assert.Equal(t, Wow.Address, newAddress)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := SetupTestDB(t)
	assert.NoError(t, err)
	defer db.Close()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	store := ParcelStore{db}
	parcel := getTestParcel()
	id, err := store.Add(parcel)
	assert.NoError(t, err)
	assert.Greater(t, id, 0)

	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	err = store.SetStatus(id, ParcelStatusRegistered)
	assert.NoError(t, err)

	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	Wow, err := store.Get(id)
	assert.NoError(t, err)
	assert.Equal(t, Wow.Status, ParcelStatusRegistered)

}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := SetupTestDB(t)
	assert.NoError(t, err)
	defer db.Close()

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}
	store := ParcelStore{db}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		assert.NoError(t, err)
		// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	assert.NoError(t, err)
	assert.Equal(t, len(parcels), len(storedParcels))
	// получите список посылок по идентификатору клиента, сохранённого в переменной client
	// убедитесь в отсутствии ошибки
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных

	// check
	// check
	for _, parcel := range storedParcels {
		// Проверяем, что посылка с таким номером есть в parcelMap
		expectedParcel, exists := parcelMap[parcel.Number]
		assert.True(t, exists, "Посылка с номером %d не найдена в ожидаемых данных", parcel.Number)

		// Проверяем, что все поля совпадают
		assert.Equal(t, client, parcel.Client)
		assert.Equal(t, expectedParcel.Address, parcel.Address)
		assert.Equal(t, expectedParcel.Status, parcel.Status)
		assert.Equal(t, expectedParcel.CreatedAt, parcel.CreatedAt)
	}
	// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
	// убедитесь, что все посылки из storedParcels есть в parcelMap
	// убедитесь, что значения полей полученных посылок заполнены верно
}
