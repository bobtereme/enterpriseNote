package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

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
	router.HandleFunc("/home/{UserName}", getUsersNotes).Methods("GET")
	router.HandleFunc("/listUsers", listUsers).Methods("GET")
	router.HandleFunc("/listNotes", listNotes).Methods("GET")
	router.HandleFunc("/notes/search", getNote)
	router.HandleFunc("/notes/create", createANote)
	router.HandleFunc("/notes/analyse/{ID}", analyseANote)
	router.HandleFunc("/notes/update/{ID}", updateNote)
	router.HandleFunc("/notes/share/{ID}", shareNote)
	router.HandleFunc("/notes/delete/{ID}", deleteNote)
	router.HandleFunc("/notes/viewPrivileges/{ID}", viewPrivileges)
	router.HandleFunc("/notes/createSharedSetting/{ID}", saveSharedSettingOnNote)
	router.HandleFunc("/notes/editPrivileges/{ID}", editPrivileges)
	router.HandleFunc("/signUp", signUp).Methods("GET")
	router.HandleFunc("/createUser", createUser).Methods("POST")
	router.HandleFunc("/logout", logout)
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
	//getUsersNotes(w, r)

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
	//fmt.Println(r)
	cookie := checkLoggedIn(r)
	//.Println(cookie.Value)

	if cookie != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	t, err := template.ParseFiles("logintemplate.html")
	if err != nil {
		log.Fatal(err)
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
					Name:  "logged-in", // cookie name
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

	} else {
		err = t.Execute(w, nil)
		if err != nil {
			log.Fatal(err)
		}
	}

}

func getUsersNotes(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	//fmt.Println(params)
	cookie := checkLoggedIn(r)
	//fmt.Println(cookie)
	//fmt.Println(cookie.Value)
	if cookie == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if cookie.Value == params["UserName"] {
		t, err := template.ParseFiles("homepageTemplate.html")
		if err != nil {
			log.Fatal(err)
		}
		userNotes := getUsersNotesDB(params["UserName"])
		err = t.Execute(w, userNotes)
		if err != nil {
			log.Fatal(err)

		}
	} else {
		http.Redirect(w, r, "/LogIn", http.StatusSeeOther)
	}
}
func getUsersNotesDB(paramater string) []Note {
	db := connectDatabase()

	rows, err := db.Query("SELECT DISTINCT _note.note_id,_note.note_owner,_note.title,_note.body,_note.date_created FROM _note LEFT JOIN _note_privileges ON _note.note_id = _note_privileges.note_id WHERE _note.note_owner = '" + paramater + "' OR (_note_privileges.user_name = '" + paramater + "' AND _note_privileges.read = true)")
	if err != nil {
		log.Fatal(err)
	}

	var userNotes []Note
	var note Note
	for rows.Next() {
		err = rows.Scan(&note.ID, &note.NoteOwner, &note.Title, &note.Body, &note.DateCreated)
		if err != nil {
			log.Fatal(err)
		}
		userNotes = append(userNotes, note)
	}
	//fmt.Println(userNotes)
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

	cookie := checkLoggedIn(r)
	if cookie == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	t, err := template.ParseFiles("listUsersTemplate.html")
	if err != nil {
		log.Fatal(err)
	}

	users := getUsersDB()

	err = t.Execute(w, users)
	if err != nil {
		log.Fatal(err)
	}

}

func getUsersDB() []User {
	db := connectDatabase()
	rows, err := db.Query(`SELECT user_name, given_name, family_name FROM "_user"`)
	if err != nil {
		log.Fatal(err)
	}

	var users []User
	var user User

	for rows.Next() {
		err = rows.Scan(&user.UserName, &user.GivenName, &user.FamilyName)
		if err != nil {
			log.Fatal(err)
		}
		users = append(users, user)
	}
	return users
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

	cookie := checkLoggedIn(r)
	if cookie == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	t, err := template.ParseFiles("listAllNotes.html")
	if err != nil {
		log.Fatal(err)
	}

	notes := listNotesDB()

	err = t.Execute(w, notes)
	if err != nil {
		log.Fatal(err)
	}

}

func listNotesDB() []Note {
	db := connectDatabase()
	rows, err := db.Query(`SELECT * FROM "_note"`)
	if err != nil {
		log.Fatal(err)
	}

	var notes []Note
	var note Note

	for rows.Next() {
		err = rows.Scan(&note.ID, &note.NoteOwner, &note.Title, &note.Body, &note.DateCreated)
		if err != nil {
			log.Fatal(err)
		}
		notes = append(notes, note)
	}
	return notes
}

func analyseANote(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Hello")
	count := 0
	params := mux.Vars(r)
	cookie := checkLoggedIn(r)
	//fmt.Println("hello")
	if cookie == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return

	}

	t, err := template.ParseFiles("analyseANoteTemplate.html")
	//fmt.Println("Hello")
	if err != nil {
		log.Fatal(err)
	}
	if r.Method == http.MethodPost {
		count = analyseANoteDB(r.FormValue("search"), params["ID"])
	}

	err = t.Execute(w, struct {
		ID    string
		Count int
	}{params["ID"], count})
	if err != nil {
		log.Fatal(err)
	}

}

func analyseANoteDB(searchInput string, id string) int {
	count := 0
	input := searchInput
	var body string
	db := connectDatabase()
	rows, err := db.Query("SELECT _note.body FROM _note WHERE _note.note_id =" + id)
	if err != nil {
		log.Fatal(err)
		return 0
	}

	for rows.Next() {
		err = rows.Scan(&body)
		if err != nil {
			log.Fatal(err)
			return 0
		}
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
		return 0
	}

	count = strings.Count(body, input)
	return count
}
func getNote(w http.ResponseWriter, r *http.Request) {
	//params := mux.Vars(r)
	cookie := checkLoggedIn(r)
	//fmt.Println(cookie.Value)

	if cookie == nil {

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	t, err := template.ParseFiles("searchNoteTemplate.html")
	if err != nil {
		log.Fatal(err)
	}
	var searchNotes []Note
	if r.Method == http.MethodPost {
		fmt.Println(cookie)
		searchNotes = searchDB(r.FormValue("search"), cookie.Value)
	}

	err = t.Execute(w, searchNotes)
	if err != nil {
		log.Fatal(err)
	}

}

func searchDB(textPattern string, user_name string) []Note {

	var searchNotes []Note
	var input = textPattern
	var note Note
	fmt.Println(input)
	fmt.Println(user_name)
	db := connectDatabase()
	rows, err := db.Query("SELECT _note.note_id,_note.note_owner,_note.title,_note.body,_note.date_created FROM _note LEFT JOIN _note_privileges ON _note.note_id = _note_privileges.note_id WHERE (_note.note_owner =  '" + user_name + "'  OR (_note_privileges.user_name =  '" + user_name + "'  AND _note_privileges.read = true)) AND _note.body LIKE " + "'%" + textPattern + "%'" + " OR _note.title LIKE " + "'%" + textPattern + "%'")
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {

		err = rows.Scan(&note.ID, &note.NoteOwner, &note.Title, &note.Body, &note.DateCreated)
		if err != nil {
			log.Fatal(err)
		}
		searchNotes = append(searchNotes, note)

	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return searchNotes

	// return true if user name exists
}

func createANote(w http.ResponseWriter, r *http.Request) {
	cookie := checkLoggedIn(r)

	if cookie == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	t, err := template.ParseFiles("createNoteTemplate.html")
	if err != nil {
		log.Fatal(err)
	}
	info := newNoteSelectDB(cookie.Value)
	if r.Method == http.MethodPost {
		newNoteInsertDB(cookie.Value, r.FormValue("title"), r.FormValue("body"), r.FormValue("settingSelect"))
		http.Redirect(w, r, "/home/"+cookie.Value, http.StatusSeeOther)
	}

	err = t.Execute(w, info)
	if err != nil {
		log.Fatal(err)
	}
}

func newNoteSelectDB(username string) []ShareSetting {
	var settings []ShareSetting
	var setting ShareSetting
	db := connectDatabase()
	rows, err := db.Query("SELECT DISTINCT name FROM _share_settings WHERE ownername = '" + username + "'")
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		err = rows.Scan(&setting.Name)
		if err != nil {
			log.Fatal(err)
		}
		settings = append(settings, setting)
	}
	return settings

}

func newNoteInsertDB(username string, title string, body string, selectSetting string) bool {

	var newNote Note

	newNote.NoteOwner = username

	newNote.Title = title
	newNote.Body = body
	newNote.DateCreated = time.Now()

	query := "INSERT INTO _note(note_owner, title,body,date_created) VALUES ($1,$2,$3,$4) RETURNING note_id;"
	db := connectDatabase()
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Fatal(err)
		return false
	}

	var id string
	err = stmt.QueryRow(newNote.NoteOwner, newNote.Title, newNote.Body, newNote.DateCreated).Scan(&id)
	if err != nil {
		log.Fatal(err)
		return false
	}
	newNote.ID = id

	selectedSetting := selectSetting

	var setting ShareSetting

	rows, err := db.Query(" SELECT _share_settings.shareduserid,_share_settings.read, _share_settings.write FROM _share_settings WHERE ownername = '" + username + "' AND _share_settings.name = '" + selectedSetting + "'")
	if err != nil {
		log.Fatal(err)
		return false
	}
	for rows.Next() {
		err = rows.Scan(&setting.ShareUserName, &setting.Read, &setting.Write)
		if err != nil {
			log.Fatal(err)
			return false
		}

		query := "INSERT INTO _note_privileges (note_id,user_name,read,write) VALUES( $1,$2,$3,$4)"
		stmt, err := db.Prepare(query)
		if err != nil {
			log.Fatal(err)
		}
		_, err = stmt.Exec(id, setting.ShareUserName, setting.Read, setting.Write)
		if err != nil {
			log.Fatal(err)
			return false
		}
	}
	return true
}
func isOwner(w http.ResponseWriter, r *http.Request) bool {
	params := mux.Vars(r)
	cookie := checkLoggedIn(r)
	if cookie == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return false
	}

	userValue := isOwnerDB(params["ID"], cookie.Value)

	if userValue != cookie.Value {
		http.Redirect(w, r, "/home/"+cookie.Value, http.StatusSeeOther)
		return false
	}
	return true
}
func isOwnerDB(note_id string, userID string) string {
	var userValue string
	db := connectDatabase()
	rows, err := db.Query("SELECT note_owner FROM _note WHERE _note.note_id = '" + note_id + "' AND _note.note_owner = '" + userID + "'")
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {

		err = rows.Scan(&userValue)
		if err != nil {
			log.Fatal(err)
		}
	}
	return userValue
}
func shareNote(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	cookie := checkLoggedIn(r)
	if cookie == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if isOwner(w, r) {
		t, err := template.ParseFiles("shareNote.html")
		if err != nil {
			log.Fatal(err)
		}
		if r.Method == http.MethodPost {
			if r.FormValue("username") != "" {
				shareNoteDB(r.FormValue("username"), r.FormValue("readAccess"), r.FormValue("writeAccess"), params["ID"])
				http.Redirect(w, r, "/home/"+cookie.Value, http.StatusSeeOther)
			} else {
				http.Redirect(w, r, "notes/share/"+params["ID"], http.StatusSeeOther)
			}

		}

		err = t.Execute(w, nil)
		if err != nil {
			log.Fatal(err)
		}

	}
}

func shareNoteDB(username string, read string, write string, id string) bool {
	var newNotePrivileges NotePriviliges

	newNotePrivileges.UserName = username
	newNotePrivileges.ID = id
	readValue := read
	if readValue == "on" {
		newNotePrivileges.Read = true
	} else {
		newNotePrivileges.Read = false
	}

	writeValue := write
	if writeValue == "on" {
		newNotePrivileges.Write = true
		newNotePrivileges.Read = true
	} else {
		newNotePrivileges.Write = false
	}

	db := connectDatabase()
	query := "INSERT INTO _note_privileges(user_name,note_id,read,write)VALUES($1,$2,$3,$4)"
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Fatal(err)
		return false
	}
	_, err = stmt.Exec(newNotePrivileges.UserName, newNotePrivileges.ID, newNotePrivileges.Read, newNotePrivileges.Write)
	if err != nil {
		log.Fatal(err)
		return false
	}
	return true
}
func updateNote(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	cookie := checkLoggedIn(r)
	if cookie == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	writeValue, note := selectNoteUpdateDB(params["ID"])
	id := note.ID

	if id != cookie.Value && writeValue == false {
		http.Redirect(w, r, "/home/"+cookie.Value, http.StatusSeeOther)
	}

	t, err := template.ParseFiles("updateNoteTemplate.html")
	if err != nil {
		log.Fatal(err)
	}
	if r.Method == http.MethodPost {
		InsertUpdateNoteDB(r.FormValue("title"), r.FormValue("body"), params["ID"])
		http.Redirect(w, r, "/home/"+cookie.Value, http.StatusSeeOther)
	}
	err = t.Execute(w, note)
	if err != nil {
		log.Fatal(err)
	}
}

func selectNoteUpdateDB(note_id string) (bool, Note) {

	var writeValue bool
	var note Note
	db := connectDatabase()
	rows, err := db.Query("SELECT _note_privileges.write FROM _note_privileges WHERE _note_privileges.note_id = '" + note_id + "'")
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		err = rows.Scan(&writeValue)
		if err != nil {
			log.Fatal(err)
		}
	}

	noteRow, err := db.Query("SELECT _note.note_owner,_note.title,_note.body FROM _note WHERE _note.note_id = '" + note_id + "'")

	for noteRow.Next() {
		err = noteRow.Scan(&note.ID, &note.Title, &note.Body)
		if err != nil {
			log.Fatal(err)
		}
	}
	return writeValue, note
}

func InsertUpdateNoteDB(title string, body string, note_id string) bool {
	var newNote Note
	db := connectDatabase()
	newNote.Title = title
	newNote.Body = body
	query := "UPDATE _note SET title = $1, body = $2 WHERE _note.note_id= '" + note_id + "'"
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Fatal(err)
		return false
	}
	_, err = stmt.Exec(newNote.Title, newNote.Body)
	if err != nil {
		log.Fatal(err)
		return false
	}
	return true
}

func deleteNote(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	cookie := checkLoggedIn(r)
	if cookie == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if isOwner(w, r) {
		deleteNoteDB(params["ID"])
		http.Redirect(w, r, "/home/"+cookie.Value, http.StatusSeeOther)

	}

}

func deleteNoteDB(note_id string) bool {
	db := connectDatabase()
	_, err := db.Exec("DELETE FROM _note_privileges WHERE _note_privileges.note_id='" + note_id + "'")
	if err != nil {
		log.Fatal(err)
		return false
	}

	_, err = db.Exec("DELETE FROM _note WHERE _note.note_id='" + note_id + "'")
	if err != nil {
		log.Fatal(err)
		return false
	}
	return true
}
func logout(w http.ResponseWriter, r *http.Request) {
	cookie := checkLoggedIn(r)
	if cookie == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "logged-in",
		MaxAge:  -1,
		Expires: time.Now().Add(-100 * time.Hour), // Set expires for older versions of IE
		Path:    "/",
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func viewPrivileges(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	if isOwner(w, r) {
		t, err := template.ParseFiles("viewPrivileges.html")
		matches := privilegesDB(params["ID"])
		err = t.Execute(w, matches)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func privilegesDB(note_id string) []NotePriviliges {
	db := connectDatabase()
	matching, err := db.Query("SELECT np.user_name,np.note_id,np.read,np.write as np FROM _note_privileges AS np INNER JOIN _note ON np.note_id = _note.note_id WHERE _note.note_id ='" + note_id + "' AND np.read = true")
	if err != nil {
		log.Fatal(err)
	}

	var matches []NotePriviliges
	var note NotePriviliges

	for matching.Next() {

		err = matching.Scan(&note.UserName, &note.ID, &note.Read, &note.Write)
		if err != nil {
			log.Fatal(err)
		}

		matches = append(matches, note)
	}
	return matches
}

func editPrivileges(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	cookie := checkLoggedIn(r)
	if cookie == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if isOwner(w, r) {
		t, err := template.ParseFiles("editPrivileges.html")
		if err != nil {
			log.Fatal(err)
		}
		if r.Method == http.MethodPost {
			editPrivilegesDB(r.FormValue("readAccess"), r.FormValue("writeAccess"), params["ID"])
			http.Redirect(w, r, "/home/"+cookie.Value, http.StatusSeeOther)

		}

		err = t.Execute(w, nil)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func editPrivilegesDB(read string, write string, note_id string) bool {
	var newNotePrivileges NotePriviliges

	if read == "on" {
		newNotePrivileges.Read = true
	} else {
		newNotePrivileges.Read = false
	}

	if write == "on" {
		newNotePrivileges.Write = true
		newNotePrivileges.Read = true
	} else {
		newNotePrivileges.Write = false
	}
	db := connectDatabase()
	query := "UPDATE _note_privileges SET read = $1, write = $2 WHERE _note_privileges.note_id ='" + note_id + "'"
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Fatal(err)
		return false
	}

	_, err = stmt.Exec(newNotePrivileges.Read, newNotePrivileges.Write)
	if err != nil {
		log.Fatal(err)
		return false
	}
	return true

}

func saveSharedSettingOnNote(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	cookie := checkLoggedIn(r)
	if cookie == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	t, err := template.ParseFiles("saveShareSettings.html")
	if err != nil {
		log.Fatal(err)
	}

	if r.Method == "POST" {
		fmt.Println("Hello")
		saveSharedSettingOnNoteDB(r.FormValue("settingName"), params["ID"])
		http.Redirect(w, r, "/home/"+cookie.Value, http.StatusSeeOther)
	}
	err = t.Execute(w, nil)
	if err != nil {
		log.Fatal(err)
	}
}
func saveSharedSettingOnNoteDB(settingName string, noteID string) bool {
	var setting ShareSetting

	setting.Name = settingName
	db := connectDatabase()
	rows, err := db.Query("SELECT n.note_owner as owner, na.user_name, na.read, na.write FROM _note_privileges AS na INNER JOIN _note as n ON na.note_id = n.note_id WHERE n.note_id = '" + noteID + "'")

	if err != nil {
		log.Fatal(err)
		return false
	}
	for rows.Next() {

		err = rows.Scan(&setting.OwnerName, &setting.ShareUserName, &setting.Read, &setting.Write)
		if err != nil {
			log.Fatal(err)
			return false
		}
		db := connectDatabase()
		query := `INSERT INTO _share_settings (ownername, shareduserid, read, write, name) VALUES ($1, $2, $3, $4, $5)`
		stmt, err := db.Prepare(query)
		if err != nil {
			log.Fatal(err)
			return false
		}
		_, err = stmt.Exec(setting.OwnerName, setting.ShareUserName, setting.Read, setting.Write, setting.Name)
		if err != nil {
			log.Fatal(err)
			return false
		}
	}
	return true
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
	ID          string    `json:"note_id:`
	NoteOwner   string    `json:"note_owner"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	DateCreated time.Time `json:"date_created"`
}
type NotePriviliges struct {
	NotePriviligesID int    `json: note_privileges_id`
	ID               string `json: note_id`
	UserName         string `json: user_name`
	Read             bool   `json: read`
	Write            bool   `json: write`
}

type ShareSetting struct {
	SharedSettingsID int
	OwnerName        string
	ShareUserName    string
	Read             bool
	Write            bool
	Name             string
}
