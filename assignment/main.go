package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func connectDatabase() (db *sql.DB) {
	//Open db connection
	db, err := sql.Open("postgres", "user=postgres password=password dbname=noteDB sslmode=disable")

	if err != nil {
		log.Panic(err)
	}

	return db
}

func main() {

	router := mux.NewRouter().StrictSlash(true)

	// mock data - @todo implement db

	router.HandleFunc("/login", login).Methods("GET")
	router.HandleFunc("/login/listUsers", listUsers).Methods("GET")
	router.HandleFunc("/notes", listNotes).Methods("GET")
	router.HandleFunc("/notes/search", getNote).Methods("GET")
	router.HandleFunc("/notes/create", createNote).Methods("POST")
	router.HandleFunc("/notes/update/{id}", updateNote).Methods("PUT")
	router.HandleFunc("/notes/delete/{id}", deleteNote).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8080", router))
}

// LOGIN
func login(w http.ResponseWriter, r *http.Request) {

	db := connectDatabase()
	defer db.Close()

	var userLogin User
	req, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(req, &userLogin)

	if checkUsername(userLogin.UserName) {

		usernameCookie := &http.Cookie{ // create a username cookie
			Name:  "username",         // cookie name
			Value: userLogin.UserName, // stored username
		}

		http.SetCookie(w, usernameCookie)                                    // set user name cookie
		fmt.Fprint(w, "Login Successfull, logged in as "+userLogin.UserName) // print for correct login details

	} else {
		fmt.Fprint(w, "Login Unsuccessfull, bad username") // print for incorrect login details
	}

}

func listUsers(w http.ResponseWriter, r *http.Request) {

	users := make([]User, 0)

	db := connectDatabase() // db connection
	defer db.Close()        // close db connection after use

	rows, err := db.Query("Select * FROM _user") // sql query sent to db $1 is the user name
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		user := User{}
		err := rows.Scan(&user.GivenName, &user.FamilyName, &user.UserName, &user.Password)
		if err != nil {
			log.Fatal(err)
		}
		users = append(users, user)
		fmt.Fprint(w, user.GivenName+" "+user.FamilyName+"\n")

	}
	if err = rows.Err(); err != nil {
		log.Fatal(err)
		fmt.Fprint(w, "it didnt work")

	}

}

// LOGIN FUNC CHECK USERNAME
func checkUsername(username string) bool {
	var name string

	db := connectDatabase() // db connection
	defer db.Close()        // close db connection after use

	getUserName, err := db.Prepare("Select user_name FROM _user WHERE user_name = $1;") // sql query sent to db $1 is the user name
	if err != nil {
		log.Fatal(err)
	}
	err = getUserName.QueryRow(username).Scan(&name) // sending query

	if err == sql.ErrNoRows {
		return false // return false if username does not exist
	}
	if err != nil {
		log.Fatal(err.Error())
	}

	return true // return true if user name exists
}
func listNotes(w http.ResponseWriter, r *http.Request) {

	notes := make([]Note, 0)

	db := connectDatabase() // db connection
	defer db.Close()        // close db connection after use

	rows, err := db.Query("Select * FROM _note") // sql query sent to db $1 is the user name
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		note := Note{}
		err := rows.Scan(&note.ID, &note.NoteOwner, &note.Title, &note.Body, &note.DateCreated)
		if err != nil {
			log.Fatal(err)
		}
		notes = append(notes, note)
		fmt.Fprint(w, note.ID+" "+note.NoteOwner+" "+note.Title+" "+note.Body+"\n")

	}
	if err = rows.Err(); err != nil {
		log.Fatal(err)
		fmt.Fprint(w, "it didnt work")

	}

}

func getNote(w http.ResponseWriter, r *http.Request) {

	db := connectDatabase()
	defer db.Close()

	var note Note
	req, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(req, &note)

	if checkNote(note.Title) {

		noteCookie := &http.Cookie{ // create a username cookie
			Name:  "title",    // cookie name
			Value: note.Title, // stored username
		}

		http.SetCookie(w, noteCookie)                      // set user name cookie
		fmt.Fprint(w, "note found with title "+note.Title) // print for correct login details

	} else {
		fmt.Fprint(w, "Note not found, bad title") // print for incorrect login details
	}

}

func checkNote(noteTitle string) bool {

	db := connectDatabase() // db connection
	defer db.Close()        // close db connection after use

	getNoteTitle, err := db.Prepare("Select * FROM _note WHERE title = $1;") // sql query sent to db $1 is the user name

	if err != nil {
		log.Fatal(err)
	}
	note := Note{}
	err = getNoteTitle.QueryRow(noteTitle).Scan(&note.ID, &note.Title, &note.NoteOwner, &note.Body, &note.DateCreated) // sending query

	switch {
	case err == sql.ErrNoRows:

		return false
	case err != nil:
		log.Fatal(err.Error())
		return false

	}
	return true
	// return true if user name exists
}

/* func listNotes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
} */
func searchNote(w http.ResponseWriter, r *http.Request) {
	db := connectDatabase() // db connection
	defer db.Close()
	title := r.FormValue("title")
	row := db.QueryRow("SELECT * FROM _note WHERE title = $1", title)
	note := Note{}
	err := row.Scan(&note.ID, &note.NoteOwner, &note.Title, &note.Body, &note.DateCreated)
	switch {
	case err == sql.ErrNoRows:
		log.Fatal(err)
		return
	case err != nil:
		log.Fatal(err)
		return
	}
	fmt.Fprint(w, "note found")

}

func createNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var note Note
	_ = json.NewDecoder(r.Body).Decode(&note)
	note.ID = strconv.Itoa(rand.Intn(10000000)) // mock ID-not safe
	notes = append(notes, note)
	json.NewEncoder(w).Encode(note)

}
func updateNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for index, item := range notes {
		if item.ID == params["id"] {
			notes = append(notes[:index], notes[index+1:]...)
			var note Note
			_ = json.NewDecoder(r.Body).Decode(&note)
			note.ID = params["id"] // mock ID-not safe
			notes = append(notes, note)
			json.NewEncoder(w).Encode(note)
			return
		}

	}
	json.NewEncoder(w).Encode(notes)
}
func deleteNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for index, item := range notes {
		if item.ID == params["id"] {
			notes = append(notes[:index], notes[index+1:]...)
			break
		}

	}
	json.NewEncoder(w).Encode(notes)
}

// MODEL USER
type User struct {
	GivenName  string `json:"givenName"`
	FamilyName string `json:"familyName"`
	UserName   string `json:"userName"`
	Password   string `json:"pasword"`
}

//modell note
type Note struct {
	ID          string `json:"note_id:`
	NoteOwner   string `json:"note_owner"`
	Title       string `json:"title"`
	Body        string `json:"body"`
	DateCreated string `json:"date_created"`
}

// init notes var as slice note struct
var notes []Note
var users []User
