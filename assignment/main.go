package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type BasicTemplate struct {
	Title string
	News  string
}

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
	router.HandleFunc("/", index).Methods("GET")
	router.HandleFunc("/login", login)
	router.HandleFunc("/home/{UserName}", home).Methods("GET")
	router.HandleFunc("/listUsers", listUsers).Methods("GET")
	router.HandleFunc("/listNotes", listNotes).Methods("GET")
	router.HandleFunc("/notes/search", getNote).Methods("GET")
	router.HandleFunc("/notes/create", createNote).Methods("POST")
	router.HandleFunc("/notes/update/{id}", updateNote).Methods("PUT")
	router.HandleFunc("/notes/delete/{id}", deleteNote).Methods("DELETE")
	router.HandleFunc("/signUp", signUp).Methods("GET")
	router.HandleFunc("/createUser", createUser).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", router))
}

func index(w http.ResponseWriter, r *http.Request) {
	p := BasicTemplate{Title: "Index", News: "Stuff"}
	t, _ := template.ParseFiles("index.html")
	t.Execute(w, p)

}
func home(w http.ResponseWriter, r *http.Request) {

	p := BasicTemplate{Title: "Home", News: "Stuff"}
	t, _ := template.ParseFiles("homepageTemplate.html")
	t.Execute(w, p)

}
func signUp(w http.ResponseWriter, r *http.Request) {
	p := BasicTemplate{Title: "Sign UP", News: "enter details:"}
	t, _ := template.ParseFiles("basictemplate.html")
	t.Execute(w, p)
}
func checkLoggedIn(r *http.Request) *http.Cookie {
	cookie, err := r.Cookie("logged-in")
	if err == http.ErrNoCookie {
		return nil
	}
	return cookie
}
func login(w http.ResponseWriter, r *http.Request) {

	cookie := checkLoggedIn(r)

	if cookie != nil {
		http.Redirect(w, r, "/index"+cookie.Value, http.StatusSeeOther)
	}

	if r.Method == http.MethodPost {

		var userLogin User
		//req, _ := ioutil.ReadAll(r.Body)
		//json.Unmarshal(req, &userLogin)
		userLogin.UserName = r.FormValue("username")
		userLogin.Password = r.FormValue("password")
		if checkUsername(userLogin.UserName) {

			cookie, err := r.Cookie("logged-in")
			if err == http.ErrNoCookie {
				cookie = &http.Cookie{ // create a username cookie
					Name:  "username", // cookie name
					Value: userLogin.UserName,
					Path:  "/", // stored username
				}
			}
			http.SetCookie(w, cookie) // set user name cookie
			//fmt.Fprint(w, "Login Successfull, logged in as "+userLogin.UserName) // print for correct login details
			http.Redirect(w, r, "/home/"+cookie.Value, http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
			//fmt.Fprint(w, "Login Unsuccessfull, bad username") // print for incorrect login details
		}

	}
	p := BasicTemplate{Title: "Login", News: "enter details:"}
	t, _ := template.ParseFiles("logintemplate.html")
	t.Execute(w, p)

}

func getUsersNotes(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	cookie := checkLoggedIn(r)
	if cookie == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if cookie.Value == params["UserName"] {
		t, err := template.ParseFiles("homepageTemplate.html")
		if err != nil {
			log.Fatal(err)
		}
		userNotes := getUsersNotesSql(params["UserName"])
		err = t.Execute(w, userNotes)
		if err != nil {
			log.Fatal(err)

		}
	} else {
		http.Redirect(w, r, "/LogIn", http.StatusSeeOther)
	}
}
func getUsersNotesSql(paramater string) []Note {
	db := connectDatabase()

	rows, err := db.Query(`SELECT DISTINCT _note.note_id,_note.note_owner,_note.title,_note.body,_note.date_created FROM note LEFT JOIN _note_privileges ON _note.note_id = _note_privileges.note_id WHERE _note.note_owner = ` + paramater + ` OR (_note_privileges._user_name = ` + paramater + ` AND _note_privileges.read = true)`)
	if err != nil {
		log.Fatal(err)
	}

	var userNotes []Note
	var note Note
	for rows.Next() {
		err = rows.Scan(&note.ID, &note.NoteOwner, &note.Title)
		if err != nil {
			log.Fatal(err)
		}
		userNotes = append(userNotes, note)
	}

	return userNotes
}
func createUser(w http.ResponseWriter, r *http.Request) {
	var newUser User
	newUser.UserName = r.FormValue("user_name")
	newUser.GivenName = r.FormValue("given_name")
	newUser.FamilyName = r.FormValue("family_name")
	newUser.Password = r.FormValue("password")

	db := connectDatabase()
	defer db.Close()

	if !checkUsername(newUser.UserName) {
		stmt, err := db.Prepare("INSERT INTO _user(user_name, given_name, family_name, password) VALUES($1,$2,$3,$4);")
		if err != nil {
			log.Fatal(err)
		}
		_, err = stmt.Exec(newUser.UserName, newUser.GivenName, newUser.FamilyName, newUser.Password)
		if err != nil {
			log.Fatal(err)
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
		fmt.Fprint(w, "Username already exists")
	}
}

// LOGIN

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

/* func searchNote(w http.ResponseWriter, r *http.Request) {
	p := BasicTemplate{Title: "Search for a note", News: "Stuff"}
	t, _ := template.ParseFiles("searchNoteTemplate.html")
	t.Execute(w, p)
} */

func getNote(w http.ResponseWriter, r *http.Request) {

	var note Note
	fmt.Println(notes)
	note.Title = r.FormValue("inputPattern")

	//req, _ := ioutil.ReadAll(r.Body)
	//json.Unmarshal(req, &note)
	for i := range notes {
		if checkNote(notes[i].Title) {

			noteCookie := &http.Cookie{ // create a username cookie
				Name:  "title",    // cookie name
				Value: note.Title, // stored username
			}

			http.SetCookie(w, noteCookie) // set user name cookie

			//(notes[i].Title)
			//fmt.Fprint(w, "note found with title "+note.Title) // print for correct login details

		} else {
			//fmt.Fprint(w, "Note not found, bad title") // print for incorrect login details
		}
	}

	p := BasicTemplate{Title: "Search for a note", News: "Stuff"}
	t, _ := template.ParseFiles("searchNoteTemplate.html")
	t.Execute(w, p)

}

func checkNote(textPattern string) bool {

	db := connectDatabase() // db connection
	defer db.Close()        // close db connection after use

	getnote, err := db.Prepare("Select * FROM _note WHERE body LIKE '%$1%';") // sql query sent to db $1 is the user name

	if err != nil {
		log.Fatal(err)
	}
	note := Note{}
	var notes []Note
	notes = append(notes, note)
	err = getnote.QueryRow(textPattern).Scan(&note.ID, &note.Title, &note.NoteOwner, &note.Body, &note.DateCreated) // sending query
	fmt.Println(note)
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

type array struct {
	notes []Note
	users []User
}

// init notes var as slice note struct
var notes []Note
var users []User
