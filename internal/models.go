package internal

import "time"

type IParser interface {
	ParseContext() (*Club, error)
	ParseEvents() (*Event, error)
	ParseInt64(str string) (int64, error)
	ParseInt16(str string) (int16, error)
	ParseTime(str string) (time.Time, error)
}

type IClubSystem interface {
	StartClub() error
}

// Club сущность игрового клуба, которая описывает его характеристики
type Club struct {
	CountTables int64                // Количество столов
	StartTime   time.Time            // Время начала работы
	FinishTime  time.Time            // Конец работы
	Price       int64                // Цена за час
	Conditions  map[string]Condition // Состояния клиентов (сел за стол 1, встал в очередь и тд)
	Queue       []string             // Очередь клиентов
	Tables      map[int64]bool       // Тут лежат занятые столы. Если стол не занят, значит его нет в мапе
	WorkTables  []WorkTable          // Массив с рабочими столами, в котором отражена статистика столов. Нумерация начинается с 1, поэтому размер массива CountTables + 1
}

// Event событие, которое происходит в клубе
type Event struct {
	Timestamp   time.Time // Время события
	Id          int16     // ID действия
	ClientName  string    // Имя клиента
	NumberTable int64     // Номер стола, если zero value, действие не связано со столом
}

// Condition состояние пользователя, которая определяет его действие (сел за стол 1, встал в очередь и тд)
type Condition struct {
	Id       int16 // Id действия
	Position int64 // Стол, за которым сидит клиент, если zero value, значит он не сидит за столом.
}

// WorkTable показатели стола для расчета прибыли.
type WorkTable struct {
	Revenue     int64         // Общий доход
	WorkingTime time.Duration // Общее рабочее время
	LastStart   time.Time     // Последнее время начала аренды стола
}
