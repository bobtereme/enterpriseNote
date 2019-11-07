package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var w http.ResponseWriter
var r *http.Request

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	connectDatabase()
	//os.Exit(m.Run())

}

func TestDatabase(t *testing.T) {
	var user User
	user.UserName = "Bob"
	user.GivenName = "Henry"
	user.FamilyName = "Tyler"
	user.Password = "password"

	var userTwo User
	userTwo.UserName = "Dragon"
	userTwo.GivenName = "Tom"
	userTwo.FamilyName = "Ford"
	userTwo.Password = "password"

	var note Note
	note.ID = "13"
	note.NoteOwner = "Bob"
	note.Title = "1st note"
	note.Body = "This is a note with random"
	note.DateCreated = time.Now()

	var notePriviliges NotePriviliges
	notePriviliges.ID = "13"
	notePriviliges.Read = true
	notePriviliges.Write = true
	notePriviliges.UserName = "Dragon"
	db := connectDatabase()

	if db != nil {
		//assert.Equal(t, "Users returned", getUsers(w, r), "Should return 'Users returned'")
		assert.NotNil(t, getUsersDB(), "Should return a list of users")
		userNotes := getUsersNotesDB("1")
		assert.NotEmpty(t, userNotes, "Should not be empty")

		assert.True(t, InsertUpdateNoteDB("New title", "new contents", "1"))
		ownerID := isOwnerDB("1", "1")
		assert.NotZero(t, ownerID, "Should not be zero")
		assert.True(t, deleteNoteDB("4"), "Should be true")
		//newUser := createUser("New", "User", "password")
		//assert.NotNil(t, newUser, "Should return a user")
		searchedNotes := searchDB("o", "1")
		assert.NotEmpty(t, searchedNotes, "Should not be empty")
		newAnalyseNote := analyseANoteDB("o", "1")
		assert.NotZero(t, newAnalyseNote, "Should not be zero")
		assert.True(t, shareNoteDB("1", "on", "on", "2"), "Should be true")
		newAccess := privilegesDB("4")
		assert.NotEmpty(t, newAccess, "Should not be empty")
		assert.True(t, editPrivilegesDB("on", "on", "2"), "Should be true")
		assert.True(t, saveSharedSettingOnNoteDB("setting 1", "4"), "Should be true")
		settings := newNoteSelectDB("3")
		assert.NotEmpty(t, settings, "Should not be empty")
		assert.True(t, newNoteInsertDB("1", "new title test", "test body text", "none"), "Should be true")
		writevalue, id := selectNoteUpdateDB("1")
		assert.True(t, writevalue, "should be true")
		assert.NotEqual(t, "", id, "id should not be empty")
	}
}

func TestCheckUsername(t *testing.T) {
	username := "123"

	expected := true
	observed := checkUsername(username)

	if observed != expected {
		t.Errorf("Expected true but returned false")
	}
}

func TestIsOwner(t *testing.T) {
	result := isOwner(w, r)

	if result == false {
		t.Errorf("Is not owner")
	}
}
