package main

import (
	"fmt"
	"log"

	"github.com/AcoStar7819/go-dbc"
)

// ChrClassesRecord describes a single row from ChrClasses.dbc (Vanilla 1.12)
type ChrClassesRecord struct {
	ID              uint32 `dbc:"column=0,type=uint32"`
	IsPlayerClass   uint32 `dbc:"column=1,type=uint32"`
	DamageBonusStat uint32 `dbc:"column=2,type=uint32"`
	PowerType       uint32 `dbc:"column=3,type=uint32"`
	PetNameToken    string `dbc:"column=4,type=string"`
	Name            string `dbc:"column=5,type=locstring"`
	FileName        string `dbc:"column=14,type=string"`
	SpellClassSet   uint32 `dbc:"column=15,type=uint32"`
	Flags           uint32 `dbc:"column=16,type=uint32"`
}

func main() {
	// Example of reading
	dbcFile, err := dbc.ReadFile("ChrClasses.dbc")
	if err != nil {
		log.Fatalf("Error reading DBC: %v", err)
	}

	var records []ChrClassesRecord
	err = dbc.UnmarshalRecords(dbcFile, &records)
	if err != nil {
		log.Fatalf("Error unmarshaling records: %v", err)
	}

	// Output the first few rows
	for i := 0; i < len(records) && i < 3; i++ {
		fmt.Printf("ID=%d IsPlayerClass=%d DamageBonusStat=%d PowerType=%d PetNameToken=%s Name=%s FileName=%s SpellClassSet=%d Flags=%d \n",
			records[i].ID,
			records[i].IsPlayerClass,
			records[i].DamageBonusStat,
			records[i].PowerType,
			records[i].PetNameToken,
			records[i].Name,
			records[i].FileName,
			records[i].SpellClassSet,
			records[i].Flags,
		)
	}

	// Example of modification and writing
	for i := range records {
		if records[i].ID == 4 {
			records[i].Name = "Swordman" // change the warrior's class name
		}
	}

	// Create a new record
	newRecord := ChrClassesRecord{
		ID:              20,
		IsPlayerClass:   1,
		DamageBonusStat: 1,
		PowerType:       1,
		PetNameToken:    "Example",
		Name:            "Example",
		FileName:        "Example",
		SpellClassSet:   1,
		Flags:           1,
	}
	records = append(records, newRecord)

	// Save file
	newFile, err := dbc.MarshalRecords(records)
	if err != nil {
		log.Fatalf("Error marshaling: %v", err)
	}

	err = dbc.WriteFile("ChrClasses.dbc", newFile)
	if err != nil {
		log.Fatalf("Error writing new DBC: %v", err)
	}
}
