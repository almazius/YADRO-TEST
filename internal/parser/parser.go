package parser

import (
	"YADRO/internal"
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Parser struct {
	Log     *log.Logger
	Scanner *bufio.Scanner
}

func NewParser(file *os.File) *Parser {
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	return &Parser{
		Log:     log.New(os.Stderr, "Parser ", log.Lshortfile|log.LstdFlags),
		Scanner: scanner,
	}
}

func (p *Parser) ParseContext() (*internal.Club, error) {
	var club internal.Club
	var err error
	//scanner := bufio.NewScanner(file)
	//scanner.Split(bufio.ScanLines)
	p.Scanner.Scan()

	text := p.Scanner.Text()
	club.CountTables, err = p.ParseInt64(text)
	if err != nil {
		return nil, err
	}

	p.Scanner.Scan()
	text = p.Scanner.Text()
	times := strings.Split(text, " ")
	club.StartTime, err = p.ParseTime(times[0])
	if err != nil {
		return nil, err
	}
	club.FinishTime, err = p.ParseTime(times[1])
	if err != nil {
		return nil, err
	}

	p.Scanner.Scan()
	text = p.Scanner.Text()

	club.Price, err = p.ParseInt64(text)
	if err != nil {
		return nil, err
	}

	return &club, nil
}

func (p *Parser) ParseEvents() (*internal.Event, error) {
	//scanner := bufio.NewScanner(file)
	//scanner.Split(bufio.ScanLines)
	var err error
	event := internal.Event{}

	p.Scanner.Scan()
	text := p.Scanner.Text()
	if text == "" {
		return nil, errors.New("file end")
	}
	fmt.Println(text)
	words := strings.Split(text, " ")
	if len(words) < 3 {
		p.Log.Print("incorrect event recording")
		//fmt.Println(text)
		return nil, errors.New("incorrect event recording")
	}
	event.Timestamp, err = p.ParseTime(words[0])
	if err != nil {
		p.Log.Print(err)
		return nil, err
	}

	event.Id, err = p.ParseInt16(words[1])
	if err != nil {
		p.Log.Print(err)
		return nil, err
	}

	if !(event.Id == 1 || event.Id == 2 || event.Id == 3 || event.Id == 4) {
		return nil, errors.New("uncorrected ID")
	}

	event.ClientName = words[2]

	if event.Id == 2 && len(words) > 3 {
		event.NumberTable, err = p.ParseInt64(words[3])
		if err != nil {
			return nil, err
		}
	}

	return &event, nil
}

//func (p *Parser) OpenFile(path string) (*os.File, error) {
//	file, err := os.Open(path)
//	if err != nil {
//		p.Log.Print(err)
//		return nil, err
//	}
//	return file, nil
//}

func (p *Parser) ParseInt64(str string) (int64, error) {
	value, err := strconv.Atoi(str)
	if err != nil {
		fmt.Println(str)
		p.Log.Print(err)
		return 0, err
	}
	return int64(value), nil
}

func (p *Parser) ParseInt16(str string) (int16, error) {
	value, err := strconv.Atoi(str)
	if err != nil {
		fmt.Println(str)
		p.Log.Print(err)
		return 0, err
	}
	return int16(value), nil
}

func (p *Parser) ParseTime(str string) (time.Time, error) {
	t, err := time.Parse("15:04", str)
	if err != nil {
		fmt.Println(str)
		p.Log.Print(err)
		return t, err
	}
	return t, nil
}
