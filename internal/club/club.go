package club

import (
	"YADRO/internal"
	"YADRO/internal/parser"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"time"
)

// ClubSystem сущность, управляющая игровым клубом
type ClubSystem struct {
	Log    *log.Logger      // Логер
	Parser internal.IParser // Парсер
}

// NewClubSystem создает экземпляр класса ClubSystem и инициализирует его
func NewClubSystem(path string) (internal.IClubSystem, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &ClubSystem{
		Log:    log.New(os.Stderr, "ClubSystem: ", log.LstdFlags),
		Parser: parser.NewParser(file),
	}, nil
}

// StartClub инициализирует начало работы клуба и сбора аналитики по нему
func (cls *ClubSystem) StartClub() error {
	var event *internal.Event
	club, err := cls.Parser.ParseContext()
	if err != nil {
		return err
	}

	club.Conditions = make(map[string]internal.Condition)
	club.Tables = make(map[int64]bool)
	club.WorkTables = make([]internal.WorkTable, club.CountTables+1)

	fmt.Println(club.StartTime.Format("15:04"))
	for true {
		event, err = cls.Parser.ParseEvents()
		if err != nil {
			if err.Error() == "file end" {
				cls.Log.Println("File end")
				break
			}
			return err
		}
		err = cls.analysisEvent(event, club)
		if err != nil {
			return err
		}
	}
	err = cls.finishClub(club)
	if err != nil {
		return err
	}
	for i := 1; i < len(club.WorkTables); i++ {
		fmt.Println(i, club.WorkTables[i].Revenue, fmt.Sprintf("%02d:%02d",
			int(club.WorkTables[i].WorkingTime.Minutes()/60), int(club.WorkTables[i].WorkingTime.Minutes())%60))
	}

	return nil
}

// analysisEvent анализирует действия клиентов и находит ошибки, если они есть.
func (cls *ClubSystem) analysisEvent(event *internal.Event, club *internal.Club) error {
	if event.Timestamp.After(club.FinishTime) || event.Timestamp.Before(club.StartTime) {
		cls.createError(event.Timestamp, "NotOpenYet", "Club already closed")
		return nil
	}
	switch event.Id {
	case 1:
		err := cls.clientCome(event, club)
		if err != nil {
			return err
		}
	case 2:
		err := cls.clientSitOnTable(event, club)
		if err != nil {
			return err
		}
	case 3:
		err := cls.clientStandOnQueue(event, club)
		if err != nil {
			return err
		}
	case 4:
		err := cls.clientLeave(event, club)
		if err != nil {
			return err
		}
	}
	return nil
}

// clientCome выполняет инструкции, если клиент пришел в клуб
func (cls *ClubSystem) clientCome(event *internal.Event, club *internal.Club) error {
	//if !(event.Timestamp.After(club.StartTime) && event.Timestamp.Before(club.FinishTime)) {
	//	cls.createError(event.Timestamp, "NotOpenYet", "user came at the wrong time")
	//} else
	if condition, exist := club.Conditions[event.ClientName]; exist &&
		(condition.Id != 4 && condition.Id != 11) {
		cls.createError(event.Timestamp, "YouShallNotPass", "client already in club")
	} else {
		club.Conditions[event.ClientName] = internal.Condition{
			Id:       event.Id,
			Position: event.NumberTable,
		}
	}
	return nil
}

// clientSitOnTable выполняет инструкции, если клиент сел за стол
func (cls *ClubSystem) clientSitOnTable(event *internal.Event, club *internal.Club) error {
	if cls.placeIsBusy(event.NumberTable, club) {
		cls.createError(event.Timestamp, "PlaceIsBusy", "place already busy")
	} else if club.Conditions[event.ClientName].Position != 0 && // проверка на zero value
		(club.Conditions[event.ClientName].Id == 2 || club.Conditions[event.ClientName].Id == 12) {
		oldTablesNumber := club.Conditions[event.ClientName].Position
		delete(club.Tables, oldTablesNumber) // освобождаю старый стол
		err := cls.finishCost(&club.WorkTables[oldTablesNumber], event.Timestamp, club.Price)
		if err != nil {
			return err
		}

		club.Conditions[event.ClientName] = internal.Condition{
			Id:       event.Id,
			Position: event.NumberTable,
		}
		club.Tables[event.NumberTable] = true // занимаю новый стол
		err = cls.startCost(&club.WorkTables[event.NumberTable], event.Timestamp)
		if err != nil {
			return err
		}
	} else if condition, exist := club.Conditions[event.ClientName]; !exist ||
		condition.Id == 4 || condition.Id == 11 {
		cls.createError(event.Timestamp, "ClientUnknown", "client didn't come")
	} else if event.NumberTable < 1 || event.NumberTable > club.CountTables {
		cls.createError(event.Timestamp, "Error", "incorrect table")
	} else {
		club.Conditions[event.ClientName] = internal.Condition{
			Id:       event.Id,
			Position: event.NumberTable,
		}
		club.Tables[event.NumberTable] = true
		club.WorkTables[event.NumberTable].LastStart = event.Timestamp
	}
	return nil
}

// clientStandOnQueue выполняет инструкции если клиент встал в очередь
func (cls *ClubSystem) clientStandOnQueue(event *internal.Event, club *internal.Club) error {
	if !cls.allTableIsBusy(club) {
		cls.createError(event.Timestamp, "ICanWaitNoLonger!", "client can sit")
	} else if int64(len(club.Queue)+1) > club.CountTables {
		cls.Log.Println("queue is too long")
		err := cls.kickClient(event, club)
		if err != nil {
			return err
		}
	} else if condition, exist := club.Conditions[event.ClientName]; !exist || condition.Id == 4 || condition.Id == 11 {
		cls.createError(event.Timestamp, "ClientUnknown", "client didn't come")
	} else if club.Conditions[event.ClientName].Position != 0 {
		cls.createError(event.Timestamp, "Error", "Client already seat on table")
	} else {
		club.Queue = append(club.Queue, event.ClientName)
		club.Conditions[event.ClientName] = internal.Condition{
			Id:       event.Id,
			Position: event.NumberTable,
		}
	}
	return nil
}

// clientLeave выполняет инструкции, если клиент ушел
func (cls *ClubSystem) clientLeave(event *internal.Event, club *internal.Club) error {
	if condition, exist := club.Conditions[event.ClientName]; !exist || condition.Id == 4 || condition.Id == 11 {
		cls.createError(event.Timestamp, "ClientUnknown", "client didn't come")
	} else {
		if club.Conditions[event.ClientName].Position != 0 {
			err := cls.finishCost(&club.WorkTables[club.Conditions[event.ClientName].Position], event.Timestamp, club.Price)
			if err != nil {
				return err
			}
			delete(club.Tables, club.Conditions[event.ClientName].Position)
			//cls.takeTable(club, event.ClientName)
			if len(club.Queue) > 0 {
				err = cls.freeTable(event, club)
				if err != nil {
					return err
				}
			}
			club.Conditions[event.ClientName] = internal.Condition{Id: 4, Position: 0}
		}
	}
	return nil
}

// finishClub реализует процесс закрытия клуба (выгнать всех клиентов, не забыв при это попросить оплату)
func (cls *ClubSystem) finishClub(club *internal.Club) error {
	kickList := cls.createKickList(club)
	sort.Strings(kickList)
	for i := range kickList {
		err := cls.kickClient(&internal.Event{
			Timestamp:   club.FinishTime,
			Id:          11,
			ClientName:  kickList[i],
			NumberTable: 0,
		}, club)
		if err != nil {
			return err
		}
	}
	fmt.Println(club.FinishTime.Format("15:04"))
	return nil
}

// createKickList создает список людей, которые находятся в клубе в момент закрытия
func (cls *ClubSystem) createKickList(club *internal.Club) []string {
	kickList := make([]string, 0, len(club.Tables))
	for clientName, condition := range club.Conditions {
		if condition.Id != 11 && condition.Id != 4 {
			if condition.Position != 0 {
				delete(club.Tables, condition.Position)
			}
			kickList = append(kickList, clientName)
		}
	}
	return kickList
}

// freeTable при появлении свободного стола, позволяет первому из очереди занять его
func (cls *ClubSystem) freeTable(event *internal.Event, club *internal.Club) error {
	oldTablesNumber := club.Conditions[event.ClientName].Position
	fmt.Println(event.Timestamp.Format("15:04"), 12, club.Queue[0], oldTablesNumber)
	club.Tables[club.Conditions[event.ClientName].Position] = true
	err := cls.startCost(&club.WorkTables[oldTablesNumber], event.Timestamp)
	if err != nil {
		return err
	}
	club.Conditions[club.Queue[0]] = internal.Condition{
		Id:       12,
		Position: club.Conditions[event.ClientName].Position,
	}

	if len(club.Queue) != 1 {
		club.Queue = club.Queue[1:]
	} else {
		club.Queue = make([]string, 0)
	}
	return nil
}

// createError создание ошибки
func (cls *ClubSystem) createError(timestamp time.Time, message, err string) {
	fmt.Println(timestamp.Format("15:04"), "13", message)
	cls.Log.Println(err)
}

// placeIsBusy проверяет, занято ли место
func (cls *ClubSystem) placeIsBusy(numberTable int64, club *internal.Club) bool {
	_, exist := club.Tables[numberTable]
	return exist
}

// allTableIsBusy проверяет, все ли места заняты
func (cls *ClubSystem) allTableIsBusy(club *internal.Club) bool {
	return int64(len(club.Tables)) == club.CountTables
}

// kickClient выгоняет клиента
func (cls *ClubSystem) kickClient(event *internal.Event, club *internal.Club) error {
	err := cls.finishCost(&club.WorkTables[club.Conditions[event.ClientName].Position], event.Timestamp, club.Price)
	if err != nil {
		return err
	}
	club.Conditions[event.ClientName] = internal.Condition{
		Id:       11,
		Position: 0,
	}
	fmt.Println(event.Timestamp.Format("15:04"), 11, event.ClientName)
	return nil
}

// finishCost считает стоимость аренды стола при окончании аренды
func (cls *ClubSystem) finishCost(table *internal.WorkTable, finishTime time.Time, price int64) error {
	if table == nil {
		cls.Log.Println("table is nil")
		return errors.New("table is nil")
	}
	table.Revenue += int64(math.Ceil((finishTime.Sub(table.LastStart)).Minutes()/60)) * price
	table.WorkingTime += finishTime.Sub(table.LastStart)
	return nil
}

// startCost запоминает время, когда клиент сел за стол, чтобы далее,
// учитывая эти данные, можно было посчитать стоимость аренды
func (cls *ClubSystem) startCost(table *internal.WorkTable, startTime time.Time) error {
	if table == nil {
		cls.Log.Println("table is nil")
		return errors.New("table is nil")
	}
	table.LastStart = startTime
	return nil
}
