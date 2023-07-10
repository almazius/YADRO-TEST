package club

import (
	"YADRO/internal"
	"YADRO/internal/parser"
	"fmt"
	"log"
	"os"
	"sort"
	"time"
)

type ClubSystem struct {
	Log    *log.Logger
	Parser *parser.Parser
}

func NewClubSystem(path string) (*ClubSystem, error) {
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

func (cls *ClubSystem) Start() error {
	var event *internal.Event
	club, err := cls.Parser.ParseContext()
	if err != nil {
		return err
	}
	club.Conditions = make(map[string]internal.Condition)
	club.Tables = make(map[int64]bool)

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
		err = cls.AnalysisEvent(event, club)
		if err != nil {
			return err
		}
	}
	cls.FinishClub(club)

	return nil
}

func (cls *ClubSystem) AnalysisEvent(event *internal.Event, club *internal.Club) error {
	if event.Id == 1 {
		if !(event.Timestamp.After(club.StartTime) && event.Timestamp.Before(club.FinishTime)) {
			cls.CreateError(event.Timestamp, "NotOpenYet", "user came at the wrong time")
		} else if condition, exist := club.Conditions[event.ClientName]; exist &&
			(condition.Id != 4 && condition.Id != 11) {
			cls.CreateError(event.Timestamp, "YouShallNotPass", "fix")
		} else {
			club.Conditions[event.ClientName] = internal.Condition{
				Id:       event.Id,
				Position: event.NumberTable,
			}
		}
	} else if event.Id == 2 {
		if cls.placeIsBusy(event.NumberTable, club) {
			cls.CreateError(event.Timestamp, "PlaceIsBusy", "fix")
		} else if club.Conditions[event.ClientName].Position != 0 && // check on zero value
			(club.Conditions[event.ClientName].Position == 2 || club.Conditions[event.ClientName].Position == 12) {
			club.Conditions[event.ClientName] = internal.Condition{
				Id:       event.Id,
				Position: event.NumberTable,
			}
			club.Tables[event.NumberTable] = true
		} else if condition, exist := club.Conditions[event.ClientName]; !exist ||
			condition.Id == 4 || condition.Id == 11 {
			cls.CreateError(event.Timestamp, "ClientUnknown", "fix")
		} else if event.NumberTable < 1 || event.NumberTable > club.CountTables {
			cls.CreateError(event.Timestamp, "Error", "incorrect table")
		} else {
			club.Conditions[event.ClientName] = internal.Condition{
				Id:       event.Id,
				Position: event.NumberTable,
			}
			club.Tables[event.NumberTable] = true
		}
	} else if event.Id == 3 {
		if !cls.allTableIsBusy(club) {
			cls.CreateError(event.Timestamp, "ICanWaitNoLonger!", "fix")
		} else if int64(len(club.Queue)+1) > club.CountTables {
			cls.Log.Println("queue is too long")
			cls.kickClient(event, club)
		} else if condition, exist := club.Conditions[event.ClientName]; !exist || condition.Id == 4 || condition.Id == 11 {
			cls.CreateError(event.Timestamp, "ClientUnknown", "fix")
		} else {
			club.Queue = append(club.Queue, event.ClientName)
			club.Conditions[event.ClientName] = internal.Condition{
				Id:       event.Id,
				Position: event.NumberTable,
			}
		}
	} else if event.Id == 4 {
		if condition, exist := club.Conditions[event.ClientName]; !exist || condition.Id == 4 || condition.Id == 11 {
			cls.CreateError(event.Timestamp, "ClientUnknown", "fix")
		} else {
			if club.Conditions[event.ClientName].Position != 0 {
				cls.takeTable(club, event.ClientName)
			}
			delete(club.Tables, club.Conditions[event.ClientName].Position)
			if len(club.Queue) > 0 {
				cls.freePC(event, club)
			}
			club.Conditions[event.ClientName] = internal.Condition{Id: 4, Position: 0}
		}
	}
	return nil
}

func (cls *ClubSystem) FinishClub(club *internal.Club) {
	kickList := cls.createKickList(club)
	sort.Strings(kickList)
	for i, _ := range kickList {
		cls.kickClient(&internal.Event{
			Timestamp:   club.FinishTime,
			Id:          11,
			ClientName:  kickList[i],
			NumberTable: 0,
		}, club)
	}
	fmt.Println(club.FinishTime.Format("15:04"))

}

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

func (cls *ClubSystem) freePC(event *internal.Event, club *internal.Club) {
	fmt.Println(event.Timestamp.Format("15:04"), 12, club.Queue[0], club.Conditions[event.ClientName].Position)
	club.Tables[club.Conditions[event.ClientName].Position] = true
	if len(club.Queue) != 1 {
		club.Queue = club.Queue[1:]
	} else {
		club.Queue = make([]string, 0)
	}
}

func (cls *ClubSystem) CreateError(timestamp time.Time, message, err string) {
	fmt.Println(timestamp.Format("15:04"), "13", message)
	cls.Log.Println(err)
}

func (cls *ClubSystem) takeTable(club *internal.Club, clientName string) {
	if len(club.Queue) > 0 {
		club.Conditions[club.Queue[0]] = internal.Condition{
			Id:       12,
			Position: club.Conditions[clientName].Position,
		}
	}
}

func (cls *ClubSystem) placeIsBusy(numberTable int64, club *internal.Club) bool {
	_, exist := club.Tables[numberTable]
	return exist
}

func (cls *ClubSystem) allTableIsBusy(club *internal.Club) bool {
	return int64(len(club.Tables)) == club.CountTables
}

func (cls *ClubSystem) kickClient(event *internal.Event, club *internal.Club) {
	club.Conditions[event.ClientName] = internal.Condition{
		Id:       11,
		Position: 0,
	}
	fmt.Println(event.Timestamp.Format("15:04"), 11, event.ClientName)
}
