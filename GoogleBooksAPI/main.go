package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

// user struct for creating session authentication
type user struct {
	UserName string
	Password []byte
	First    string
	Last     string
	Role     string
}

var tpl *template.Template
var dbUsers = map[string]user{}      // user ID, user
var dbSessions = map[string]string{} // session ID, user ID
var db *sql.DB
var initialRun = false

//Books to get JSON Data
type Books struct {
	Kind       string `json:"kind"`
	TotalItems int    `json:"totalItems"`
	Items      []struct {
		Kind       string `json:"kind"`
		ID         string `json:"id"`
		Etag       string `json:"etag"`
		SelfLink   string `json:"selfLink"`
		VolumeInfo struct {
			Title               string   `json:"title"`
			Subtitle            string   `json:"subtitle"`
			Authors             []string `json:"authors"`
			Publisher           string   `json:"publisher"`
			PublishedDate       string   `json:"publishedDate"`
			Description         string   `json:"description"`
			IndustryIdentifiers []struct {
				Type       string `json:"type"`
				Identifier string `json:"identifier"`
			} `json:"industryIdentifiers"`
			ReadingModes struct {
				Text  bool `json:"text"`
				Image bool `json:"image"`
			} `json:"readingModes"`
			PageCount           int      `json:"pageCount"`
			PrintType           string   `json:"printType"`
			Categories          []string `json:"categories"`
			MaturityRating      string   `json:"maturityRating"`
			AllowAnonLogging    bool     `json:"allowAnonLogging"`
			ContentVersion      string   `json:"contentVersion"`
			PanelizationSummary struct {
				ContainsEpubBubbles  bool `json:"containsEpubBubbles"`
				ContainsImageBubbles bool `json:"containsImageBubbles"`
			} `json:"panelizationSummary"`
			Language            string `json:"language"`
			PreviewLink         string `json:"previewLink"`
			InfoLink            string `json:"infoLink"`
			CanonicalVolumeLink string `json:"canonicalVolumeLink"`
		} `json:"volumeInfo"`
		SaleInfo struct {
			Country     string `json:"country"`
			Saleability string `json:"saleability"`
			IsEbook     bool   `json:"isEbook"`
		} `json:"saleInfo"`
		AccessInfo struct {
			Country                string `json:"country"`
			Viewability            string `json:"viewability"`
			Embeddable             bool   `json:"embeddable"`
			PublicDomain           bool   `json:"publicDomain"`
			TextToSpeechPermission string `json:"textToSpeechPermission"`
			Epub                   struct {
				IsAvailable bool `json:"isAvailable"`
			} `json:"epub"`
			Pdf struct {
				IsAvailable bool `json:"isAvailable"`
			} `json:"pdf"`
			WebReaderLink       string `json:"webReaderLink"`
			AccessViewStatus    string `json:"accessViewStatus"`
			QuoteSharingAllowed bool   `json:"quoteSharingAllowed"`
		} `json:"accessInfo"`
		SearchInfo struct {
			TextSnippet string `json:"textSnippet"`
		} `json:"searchInfo"`
	} `json:"items"`
}

//SQL Database Configuration

func configureDB() *sql.DB {
	db, err := sql.Open("mysql", "gbhatia:password@tcp(localhost:3306)/mysql?charset=utf8")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Database opened successfully")
	}

	_, err = db.Exec("CREATE DATABASE GoogleAPIDb")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Successfully created database..")
	}
	_, err = db.Exec("USE GoogleAPIDb")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("DB selected successfully..")
	}

	// cleaning prior session Data tables
	drop(db, "Books")
	drop(db, "Authors")
	drop(db, "Publishers")

	// Creating New Tables
	_, err = db.Exec("CREATE TABLE Books (BookID int NOT NULL AUTO_INCREMENT PRIMARY KEY,Title VARCHAR(100));")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Books table created..")
	}
	_, err = db.Exec("CREATE TABLE Authors (AuthorID int NOT NULL AUTO_INCREMENT PRIMARY KEY,AuthorName VARCHAR(100),BookTitle VARCHAR(100));")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Authors table created..")
	}

	_, err = db.Exec("CREATE TABLE Publishers (PublisherID int NOT NULL AUTO_INCREMENT PRIMARY KEY,PublisherName VARCHAR(100), BookTitle VARCHAR(100));")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Publisher table created..")
	}

	return db

}

//Function to delete Table from DB
func drop(db *sql.DB, tblName string) {
	s := "DROP TABLE " + tblName + ";"
	_, err := db.Exec(s)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Dropped table" + tblName)

}

func init() {
	tpl = template.Must(template.ParseGlob("templates/*"))
	db = configureDB() // Calling function to configure SQL DB at initialization
}

// function to getUser for current session, if no cookie found then create one for new user
func getUser(w http.ResponseWriter, req *http.Request) user {
	// get cookie
	c, err := req.Cookie("session")
	if err != nil {
		sID, _ := uuid.NewV4()
		c = &http.Cookie{
			Name:  "session",
			Value: sID.String(),
		}

	}
	http.SetCookie(w, c)

	// if the user exists already, get user
	var u user
	if un, ok := dbSessions[c.Value]; ok {
		u = dbUsers[un]
	}
	return u
}

// function to check if user is already looged in
func alreadyLoggedIn(req *http.Request) bool {
	c, err := req.Cookie("session")
	if err != nil {
		return false
	}
	un := dbSessions[c.Value]
	_, ok := dbUsers[un]
	return ok
}

func main() {

	http.HandleFunc("/", index)
	http.HandleFunc("/signup", signup)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/Author", GoogleBooksByAuthor) //Google API for searching Author
	http.ListenAndServe(":8080", nil)
}

func index(w http.ResponseWriter, req *http.Request) {
	u := getUser(w, req)
	tpl.ExecuteTemplate(w, "index.html", u)
}

func signup(w http.ResponseWriter, req *http.Request) {
	if alreadyLoggedIn(req) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	var u user
	// process form submission
	if req.Method == http.MethodPost {
		// get form values
		un := req.FormValue("username")
		p := req.FormValue("password")
		f := req.FormValue("firstname")
		l := req.FormValue("lastname")
		r := req.FormValue("role")
		// username taken?
		if _, ok := dbUsers[un]; ok {
			http.Error(w, "Username already taken", http.StatusForbidden)
			return
		}
		// create session
		sID, _ := uuid.NewV4()
		c := &http.Cookie{
			Name:  "session",
			Value: sID.String(),
		}
		http.SetCookie(w, c)
		dbSessions[c.Value] = un
		// store user in dbUsers
		bs, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.MinCost)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		u = user{un, bs, f, l, r}
		dbUsers[un] = u
		// redirect
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	tpl.ExecuteTemplate(w, "signup.html", u)
}

func login(w http.ResponseWriter, req *http.Request) {
	if alreadyLoggedIn(req) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	var u user
	// process form submission
	if req.Method == http.MethodPost {
		un := req.FormValue("username")
		p := req.FormValue("password")
		// is there a username?
		u, ok := dbUsers[un]
		if !ok {
			http.Error(w, "Username and/or password do not match", http.StatusForbidden)
			return
		}
		// does the entered password match the stored password?
		err := bcrypt.CompareHashAndPassword(u.Password, []byte(p))
		if err != nil {
			http.Error(w, "Username and/or password do not match", http.StatusForbidden)
			return
		}
		// create session
		sID, _ := uuid.NewV4()
		c := &http.Cookie{
			Name:  "session",
			Value: sID.String(),
		}
		http.SetCookie(w, c)
		dbSessions[c.Value] = un
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	//create()
	tpl.ExecuteTemplate(w, "login.html", u)
}

func logout(w http.ResponseWriter, req *http.Request) {
	if !alreadyLoggedIn(req) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	c, _ := req.Cookie("session")
	// delete the session
	delete(dbSessions, c.Value)
	// remove the cookie
	c = &http.Cookie{
		Name:   "session",
		Value:  "",
		MaxAge: -1,
	}
	http.SetCookie(w, c)

	http.Redirect(w, req, "/login", http.StatusSeeOther)
}

// GoogleBooksByAuthor will open the URL
func GoogleBooksByAuthor(w http.ResponseWriter, req *http.Request) {
	u := getUser(w, req) // getting current user

	if !alreadyLoggedIn(req) { // if user is not logged in then sending the user on Log In page
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	if u.Role != "admin" { // checking if user has Admin Role
		http.Error(w, "You must be Admin to search the books", http.StatusForbidden)
		return
	}

	tpl.ExecuteTemplate(w, "author.html", u) // Template for selecting Author

	// Code not to be run on initialRun
	if initialRun {
		url := "https://www.googleapis.com/books/v1/volumes?q=" + url.QueryEscape(req.FormValue("authorname"))
		res, err := http.Get(url)
		if err != nil {
			io.WriteString(w, "No Results Found")
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			io.WriteString(w, "Error reading Data")
		}
		//Converting response JSON into Books struct
		book := Books{}
		jsonErr := json.Unmarshal(body, &book)
		if jsonErr != nil {
			io.WriteString(w, "Cannot convert Data")
		}

		create(book) // Calling function Create to create SQL records

		tpl.ExecuteTemplate(w, "Display.html", book.Items) // Template to display Author response from Google API

	}

	initialRun = true

}

//function to create records in SQL
func create(b Books) {

	_, err := db.Exec("USE GoogleAPIDb")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("DB selected successfully..")
	}

	for _, bk := range b.Items {
		BookTitle := bk.VolumeInfo.Title

		//Inserting Records
		_, err = db.Exec("INSERT INTO Books (Title) VALUES (?)", BookTitle)
		if err != nil {
			panic(err)
		}
		_, err = db.Exec("INSERT INTO Publishers (PublisherName,BookTitle) VALUES (?,?)", bk.VolumeInfo.Publisher, BookTitle)
		if err != nil {
			panic(err)
		}
		for _, a := range bk.VolumeInfo.Authors {
			_, err = db.Exec("INSERT INTO Authors (AuthorName,BookTitle) VALUES (?,?)", a, BookTitle)
			if err != nil {
				panic(err)
			}
		}

	}

}
