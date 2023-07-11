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

// Parser структура, реализующая IParser, управляющая парсингом документа
type Parser struct {
	Log     *log.Logger    // Логер
	Scanner *bufio.Scanner // Сканер текста
}

func NewParser(file *os.File) internal.IParser {
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	return &Parser{
		Log:     log.New(os.Stderr, "Parser ", log.Lshortfile|log.LstdFlags),
		Scanner: scanner,
	}
}

// ParseContext парсит первые 3 строки, в которых хранятся настройки для клуба
func (p *Parser) ParseContext() (*internal.Club, error) {
	var club internal.Club
	var err error
	p.Scanner.Scan()

	text := p.Scanner.Text()
	if text == "" {
		return nil, errors.New("file uncorrected")
	}
	club.CountTables, err = p.ParseInt64(text)
	if err != nil {
		return nil, err
	}

	p.Scanner.Scan()
	text = p.Scanner.Text()
	if text == "" {
		return nil, errors.New("file uncorrected")
	}
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
	if text == "" {
		return nil, errors.New("file uncorrected")
	}

	club.Price, err = p.ParseInt64(text)
	if err != nil {
		return nil, err
	}

	return &club, nil
}

// ParseEvents парсит все события в файле
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

// ParseInt64 вспомогательная функция, позволяющая парсить числовые значения (кол-во столов)
func (p *Parser) ParseInt64(str string) (int64, error) {
	value, err := strconv.Atoi(str)
	if err != nil {
		fmt.Println(str)
		p.Log.Print(err)
		return 0, err
	}
	return int64(value), nil
}

// ParseInt16 вспомогательная функция, позволяющая парсить числовые значения (Id события)
func (p *Parser) ParseInt16(str string) (int16, error) {
	value, err := strconv.Atoi(str)
	if err != nil {
		fmt.Println(str)
		p.Log.Print(err)
		return 0, err
	}
	return int16(value), nil
}

// ParseTime вспомогательная функция, позволяющая парсить время
func (p *Parser) ParseTime(str string) (time.Time, error) {
	t, err := time.Parse("15:04", str)
	if err != nil {
		fmt.Println(str)
		p.Log.Print(err)
		return t, err
	}
	return t, nil
}
