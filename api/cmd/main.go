package main

import (
	"YADRO/internal/club"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("file not specified")
	}
	path := os.Args[1]
	cls, err := club.NewClubSystem(path)
	if err != nil {
		return
	}
	err = cls.StartClub()
	if err != nil {
		log.Fatalln(err)
	}
}
