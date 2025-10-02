/*

in-Common attributes

INPUTS:  netId
OUTPUTS: eduPersonPrimaryAffiliation, eduPersonAffiliation, eduPersonScopedAffiliation, name, preferredFirstName

This code is to return the following attributes through an web api call:

	/api/eduPersonPrimaryAffiliation - return one of the following:
		faculty,
		staff,
		student,
		alum,
		member,
		affiliate,
		employee,
		contractor,
		library-walk-in

	/api/eduPersonAffiliation - return one or more of the following:
		faculty,
		staff,
		student,
		alum,
		member,
		affiliate,
		employee,
		contractor,
		library-walk-in

	/api/eduPersonScopedAffiliation	- return one or more of the following:
		faculty@byu.edu,
		staff@byu.edu,
		student@byu.edu,
		alum@byu.edu,
		member@byu.edu,
		affiliate@byu.edu,
		employee@byu.edu,
		contractor@byu.edu
		library-walk-in@byu.edu

	/api/name	 - return the user's full name

	/api/preferredFirstName	- return the user's preferred first name

	/api/eduPerson - return all of the above attributes

*/

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

// Global database connection pool
var db *sql.DB

// This is the structure that contains all of the information. All or parts may be returned, depending on the api call.

func main() {

	// Initialize the database connection
	initDB()
	defer db.Close()
	var version string

	if err := db.QueryRow("SELECT VERSION()").Scan(&version); err != nil {
		log.Fatalf("Failed to get database version: %v", err)

	}
	log.Println("Database version: " + version)

	// Database is connected (Retrieved the version to verify connection)
	log.Println("Database connected successfully")

	// Create a new Gorilla Mux router
	gRouter := mux.NewRouter()

	// Define the routes and their handlers
	gRouter.HandleFunc("/api/eduPersonPrimaryAffiliation", getEduPersonPrimaryAffiliation).Methods("GET")
	gRouter.HandleFunc("/api/eduPersonAffiliation", getEduPersonAffiliation).Methods("GET")
	gRouter.HandleFunc("/api/eduPersonScopedAffiliation", getEduPersonScopedAffiliation).Methods("GET")
	gRouter.HandleFunc("/api/name", getName).Methods("GET")
	gRouter.HandleFunc("/api/preferredFirstName", getPreferredFirstName).Methods("GET")
	gRouter.HandleFunc("/api/eduPersonAll", getAll).Methods("GET")
	gRouter.HandleFunc("/api/help", help).Methods("GET")

	// Start the server
	log.Println("Server started on :3000")
	http.ListenAndServe(":3000", gRouter)
}

// initDB initializes the database connection
func initDB() {

	var err error
	// Open a connection to the database
	db, err = sql.Open("mysql", "media:Called2Serve1965!@tcp4(192.168.5.184:3306)/in-common?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}

	// Verify the connection to the database
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
}

func getqueryDB(query string, netId string) (string, error) {

	var result string
	err := db.QueryRow(query, netId).Scan(&result)
	if err != nil {
		return "", err
	}
	return result, nil
}

func getName(w http.ResponseWriter, r *http.Request) {
	consoleMsg := "getName called"
	log.Println(consoleMsg)

	// Get the netId from the query parameters
	netId := r.URL.Query().Get("netId")
	if netId == "" {
		http.Error(w, "Missing netId parameter", http.StatusBadRequest)
		return
	}

	// Query the database for name
	name, err := getqueryDB("SELECT name FROM users WHERE netId = ?", netId)

	if err != nil {
		http.Error(w, "Error fetching name: "+err.Error(), http.StatusInternalServerError)
		return

	}

	// Set the response header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Return the name as JSON
	response := map[string]string{"name": name}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getPreferredFirstName(w http.ResponseWriter, r *http.Request) {
	consoleMsg := "getPreferredFirstName called"
	log.Println(consoleMsg)

	// Get the netId from the query parameters
	netId := r.URL.Query().Get("netId")
	if netId == "" {
		http.Error(w, "Missing netId parameter", http.StatusBadRequest)
		return
	}

	// Query the database for preferredFirstName
	preferredFirstName, err := getqueryDB("SELECT preferredFirstName FROM users WHERE netId = ?", netId)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No user found with the given netId", http.StatusNotFound)
		} else {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Set the response header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Return the preferredFirstName as JSON
	response := map[string]string{"preferredFirstName": preferredFirstName}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get eduPersonAffiliation
// Requires netId as input
// Returns one or more of the following: faculty, staff, student, alum, member, affiliate, employee, contractor, library-walk-in
// or returns an error if the netId is not found or if there is a database error

func getEduPersonAffiliation(w http.ResponseWriter, r *http.Request) {
	consoleMsg := "getEduPersonAffiliation called"
	log.Println(consoleMsg)

	// Get the netId from the query parameters
	netId := r.URL.Query().Get("netId")
	if netId == "" {
		http.Error(w, "Missing netId parameter", http.StatusBadRequest)
		return
	}

	// Query the database for eduPersonAffiliation
	row := db.QueryRow(`
    SELECT student, faculty, staff, alum, member, affiliate, employee, primaryaffiliation
    FROM users
    WHERE netId = ?
`, netId)

	var (
		student, faculty, staff, alum, member, affiliate, employee bool
		primaryaffiliation                                         string
	)

	err := row.Scan(&student, &faculty, &staff, &alum, &member, &affiliate, &employee, &primaryaffiliation)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var affiliations []string

	if student {
		affiliations = append(affiliations, "student")
	}
	if faculty {
		affiliations = append(affiliations, "faculty")
	}
	if staff {
		affiliations = append(affiliations, "staff")
	}
	if alum {
		affiliations = append(affiliations, "alum")
	}
	if member {
		affiliations = append(affiliations, "member")
	}
	if affiliate {
		affiliations = append(affiliations, "affiliate")
	}
	if employee {
		affiliations = append(affiliations, "employee")
	}
	if primaryaffiliation == "contractor" {
		affiliations = append(affiliations, "contractor")
	}
	if primaryaffiliation == "library-walk-in" {
		affiliations = append(affiliations, "library-walk-in")
	}
	/*
		if len(affiliations) == 0 {
			http.Error(w, "No affiliations found for the given netId", http.StatusNotFound)
			return
		}
	*/
	// Set the response header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Return the eduPersonAffiliation as JSON
	response := map[string][]string{"eduPersonAffiliation": affiliations}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getEduPersonScopedAffiliation(w http.ResponseWriter, r *http.Request) {
	consoleMsg := "getEduPersonScopedAffiliation called"
	log.Println(consoleMsg)

	// Get the netId from the query parameters
	netId := r.URL.Query().Get("netId")
	if netId == "" {
		http.Error(w, "Missing netId parameter", http.StatusBadRequest)
		return
	}

	// Query the database for eduPersonScopedAffiliation
	row := db.QueryRow(`
	SELECT student, faculty, staff, alum, member, affiliate, employee, primaryaffiliation
	FROM users
	WHERE netId = ?
`, netId)
	var (
		student, faculty, staff, alum, member, affiliate, employee bool
		primaryaffiliation                                         string
	)

	err := row.Scan(&student, &faculty, &staff, &alum, &member, &affiliate, &employee, &primaryaffiliation)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var scopedAffiliations []string

	if student {
		scopedAffiliations = append(scopedAffiliations, "student@byu.edu")

	}
	if faculty {
		scopedAffiliations = append(scopedAffiliations, "faculty@byu.edu")
	}
	if staff {
		scopedAffiliations = append(scopedAffiliations, "staff@byu.edu")
	}
	if alum {
		scopedAffiliations = append(scopedAffiliations, "alum@byu.edu")
	}
	if member {
		scopedAffiliations = append(scopedAffiliations, "member@byu.edu")
	}
	if affiliate {
		scopedAffiliations = append(scopedAffiliations, "affiliate@byu.edu")
	}
	if employee {
		scopedAffiliations = append(scopedAffiliations, "employee@byu.edu")
	}
	if primaryaffiliation == "contractor" {
		scopedAffiliations = append(scopedAffiliations, "contractor@byu.edu")
	}
	if primaryaffiliation == "library-walk-in" {
		scopedAffiliations = append(scopedAffiliations, "library-walk-in@byu.edu")
	}
	/*
		if len(scopedAffiliations) == 0 {
			http.Error(w, "No scoped affiliations found for the given netId", http.StatusNotFound)
			return
		}
	*/
	// Set the response header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Return the eduPersonScopedAffiliation as JSON
	response := map[string][]string{"eduPersonScopedAffiliation": scopedAffiliations}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func getEduPersonPrimaryAffiliation(w http.ResponseWriter, r *http.Request) {
	consoleMsg := "getEduPersonPrimaryAffiliation called"
	log.Println(consoleMsg)

	// Get the netId from the query parameters
	netId := r.URL.Query().Get("netId")
	if netId == "" {
		http.Error(w, "Missing netId parameter", http.StatusBadRequest)
		return
	}

	// Query the database for eduPersonPrimaryAffiliation
	var eduPersonPrimaryAffiliation string
	err := db.QueryRow("SELECT primaryaffiliation FROM users WHERE netId = ?", netId).Scan(&eduPersonPrimaryAffiliation)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No user found with the given netId", http.StatusNotFound)
		} else {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Set the response header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Return the eduPersonPrimaryAffiliation as JSON
	response := map[string]string{"eduPersonPrimaryAffiliation": eduPersonPrimaryAffiliation}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getAll(w http.ResponseWriter, r *http.Request) {

	type UserInfo struct {
		Name                        string   `json:"name"`
		PreferredFirstName          string   `json:"preferredFirstName"`
		EduPersonPrimaryAffiliation string   `json:"eduPersonPrimaryAffiliation"`
		EduPersonAffiliation        []string `json:"eduPersonAffiliation"`
		EduPersonScopedAffiliation  []string `json:"eduPersonScopedAffiliation"`
	}
	// Create the user object
	user := UserInfo{}

	consoleMsg := "getAll called"
	log.Println(consoleMsg)

	netId := r.URL.Query().Get("netId")
	if netId == "" {
		http.Error(w, "Missing netId parameter", http.StatusBadRequest)
		return
	}

	row := db.QueryRow(`
        SELECT student, faculty, staff, alum, member, affiliate, employee, contractor, library, primaryaffiliation, name, preferredfirstname
        FROM users
        WHERE netId = ?
    `, netId)

	var (
		student, faculty, staff, alum, member, affiliate, employee, contractor, library bool
		primaryAffiliation                                                              string
	)

	err := row.Scan(&student, &faculty, &staff, &alum, &member, &affiliate, &employee, &contractor, &library, &primaryAffiliation, &user.Name, &user.PreferredFirstName)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if student {
		user.EduPersonAffiliation = append(user.EduPersonAffiliation, "student")
		user.EduPersonScopedAffiliation = append(user.EduPersonScopedAffiliation, "student@byu.edu")
	}
	if faculty {
		user.EduPersonAffiliation = append(user.EduPersonAffiliation, "faculty")
		user.EduPersonScopedAffiliation = append(user.EduPersonScopedAffiliation, "faculty@byu.edu")
	}
	if staff {
		user.EduPersonAffiliation = append(user.EduPersonAffiliation, "staff")
		user.EduPersonScopedAffiliation = append(user.EduPersonScopedAffiliation, "staff@byu.edu")
	}
	if alum {
		user.EduPersonAffiliation = append(user.EduPersonAffiliation, "alum")
		user.EduPersonScopedAffiliation = append(user.EduPersonScopedAffiliation, "alum@byu.edu")
	}
	if member {
		user.EduPersonAffiliation = append(user.EduPersonAffiliation, "member")
		user.EduPersonScopedAffiliation = append(user.EduPersonScopedAffiliation, "member@byu.edu")
	}
	if affiliate {
		user.EduPersonAffiliation = append(user.EduPersonAffiliation, "affiliate")
		user.EduPersonScopedAffiliation = append(user.EduPersonScopedAffiliation, "affiliate@byu.edu")
	}
	if employee {
		user.EduPersonAffiliation = append(user.EduPersonAffiliation, "employee")
		user.EduPersonScopedAffiliation = append(user.EduPersonScopedAffiliation, "employee@byu.edu")
	}
	if contractor {
		user.EduPersonAffiliation = append(user.EduPersonAffiliation, "contractor")
		user.EduPersonScopedAffiliation = append(user.EduPersonScopedAffiliation, "contractor@byu.edu")
	}
	if library {
		user.EduPersonAffiliation = append(user.EduPersonAffiliation, "library-walk-in")
		user.EduPersonScopedAffiliation = append(user.EduPersonScopedAffiliation, "library-walk-in@byu.edu")
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)

	// Encode struct to JSON and write the response
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
	}
}

func help(w http.ResponseWriter, r *http.Request) {
	consoleMsg := "Help called"
	log.Println(consoleMsg)

	w.Header().Set("Content-Type", "text/html")

	// Write a simple help screen
	fmt.Fprint(w, `
        <!DOCTYPE html>
        <html>
        <head>
            <title>Help Screen</title>
        </head>
        <body>
            <h1>Welcome to the Help Page</h1>
            <p>Here are some instructions to get started:</p>
            <ul>
                <li>Use <code>/api</code> to access the API.</li>
                <li>Use <code>/status</code> to check system status.</li>
                <li>Contact support at mailto:support@example.comsupport@example.com</a>.</li>
            </ul>
        </body>
        </html>
    `)

}

// Additional handler functions (getEduPersonPrimaryAffiliation, getEduPersonAffiliation, etc.) would be defined here
