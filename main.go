package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	//Package gorilla/mux implements a request router and dispatcher for matching incoming requests to their respective handler.
	//The name mux stands for HTTP request multiplexer.
	"github.com/gorilla/mux"
	//The GORM is fantastic ORM library for Golang, aims to be developer friendly. It is an ORM library for dealing with relational databases
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // not using directly in the code, need some stuff from this package
)

type Person struct {
	gorm.Model

	Name      string `json:"firstname"`
	LastName  string `json:"lastname"`
	Age       string `json:"age"`
	Birthdate string `json:"birthdate"`
}

//global variables

var db *gorm.DB
var err error

func main() {

	// Loading environments variables,because of safety issues.

	dialect := os.Getenv("DIALECT")
	host := os.Getenv("HOST")
	dbPort := os.Getenv("DBPORT")
	user := os.Getenv("USER")
	dbName := os.Getenv("NAME")
	password := os.Getenv("PASSWORD")

	// Data base connection string

	dbURI := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s port=%s", host, user, dbName, password, dbPort)

	// Openning a connectin to the DB

	db, err = gorm.Open(dialect, dbURI)

	//error handling

	if err != nil {
		log.Fatal(err)
	} else {

		fmt.Println("Successfully connected to db")
	}

	// if the db is not in use and the app is not running we want to close the connection
	// we will close when the main fucntion is finished
	// defer - do this when the current function stops running

	defer db.Close()

	// Make db migration - stumping the struct into the db	, telling the data base - thats the attribute of person,
	// do it in the postgres, we do it once.

	db.AutoMigrate(&Person{})

	// API routes

	router := mux.NewRouter()

	router.HandleFunc("/people", getPeople).Methods("GET")                   // only can get users not sending post request
	router.HandleFunc("/create/person", createPerson).Methods("POST")        // post a new user
	router.HandleFunc("/delete/person/{id}", deletePerson).Methods("DELETE") // delete a user
	router.HandleFunc("/update/person/{id}", updatePerson).Methods("PUT")    // update a user

	// http uses port 80, so 80 for http request
	//Now, when running a web server on my computer, i need to access that server somehow
	//and since port 80 is already busy, i need to use a different port to successfully connect to it.
	//Although any open port is fair game, usually such a server is configured to use port 8080.

	http.ListenAndServe(":8080", router)
}

// Api Controllers
func getPeople(w http.ResponseWriter, r *http.Request) {
	// array of people , go into the db and find all of the people , all the models who fits the struct Person
	var people []Person
	db.Find(&people)

	// convert it into json
	json.NewEncoder(w).Encode(&people)
}

func createPerson(w http.ResponseWriter, r *http.Request) {
	var person Person // creating struct to insert the information to it

	// someone sends json and we convert it into struct of Person
	json.NewDecoder(r.Body).Decode(&person) // r.Body - its the json

	createdPerson := db.Create(&person)
	err = createdPerson.Error

	// sends error in case we didnt succeed uploading to the db

	if err != nil {
		json.NewEncoder(w).Encode(err)
	} else { // in case of success , show the json that was inserted.
		json.NewEncoder(w).Encode(&person)
	}
}

func deletePerson(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r) // getting the params to extract the id .

	var person Person

	db.First(&person, params["id"]) // the first that mataches

	db.Delete(&person)

	json.NewEncoder(w).Encode(&person)

}

func updatePerson(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r) // getting the params to extract the id .

	var person Person

	db.First(&person, params["id"]) // the first that mataches

	json.NewDecoder(r.Body).Decode(&person) // r.Body - its the json

	db.Save(&person)

	json.NewEncoder(w).Encode(&person) // show the json inserted

}
