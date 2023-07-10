package internal

import "time"

type Club struct {
	CountTables int64
	StartTime   time.Time
	FinishTime  time.Time
	Price       int64
	Conditions  map[string]Condition
	Queue       []string       // очередь клиетов
	Tables      map[int64]bool // тут леэат занятые столы. Если стол не занят, значит его нет в мапе
	//Events      []Event
}

type Event struct {
	Timestamp   time.Time
	Id          int16
	ClientName  string
	NumberTable int64
}

type Condition struct {
	Id       int16
	Position int64
}
