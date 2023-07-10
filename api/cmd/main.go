package main

import (
	"YADRO/internal/club"
	"log"
)

func main() {
	cls, err := club.NewClubSystem("/home/sigy/GolandProjects/YADRO/tests/test1.txt")
	if err != nil {
		return
	}
	err = cls.Start()
	if err != nil {
		log.Fatalln(err)
	}
}
