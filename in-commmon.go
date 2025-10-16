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
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

type APICreds struct {
	APIUser string `json:"API_USER"`
	APIPass string `json:"API_PASS"`
}

type Person struct {
	netId              string   `json:"netId"`
	netIdScoped        string   `json:"netIdScoped"`
	emailAddress       string   `json:"emailAddress"`
	primaryAffiliation string   `json:"eduPersonPrimaryAffiliation"`
	affiliations       []string `json:"eduPersonAffiliation"`
	scopedAffiliations []string `json:"eduPersonScopedAffiliation"`
	name               string   `json:"name"`
	preferredFirstName string   `json:"preferredFirstName"`
}

// This is the structure that contains all of the information. All or parts may be returned, depend

func loadAPICreds(filename string) (string, string, error) {

	data, err := os.ReadFile("/Users/dwp32/APICreds/in-common.json")
	if err != nil {
		return "", "", err
	}

	var creds APICreds
	if err := json.Unmarshal(data, &creds); err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return "", "", err
	}

	return creds.APIUser, creds.APIPass, nil
}

func getToken(endpoint, user, pass string) (string, error) {
	// Build form body
	form := url.Values{}
	form.Add("grant_type", "client_credentials")

	// Create POST request with the form body
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(user, pass)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read and decode the response
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Raw response:", string(body)) // helpful for debugging

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get token: %s", resp.Status)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	// Depending on the API, the key might be "access_token" instead of "token"
	token, ok := result["access_token"].(string)
	if !ok {
		// Try alternate key just in case
		token, ok = result["token"].(string)
		if !ok {
			return "", fmt.Errorf("token not found in response")
		}
	}

	fmt.Println("Token:", token)
	return token, nil
}

func getUserBasicInfo(token, netId string) (*Person, error) {
	// Example endpoint, replace with actual
	endpoint := fmt.Sprintf("https://api-sandbox.byu.edu/byuapi/persons/v4/%s", netId)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Response Status:", resp.Status)
		return nil, fmt.Errorf("failed to get user info: %s", resp.Status)
	}

	body, _ := io.ReadAll(resp.Body)
	//fmt.Println("Raw user info response:", string(body)) // helpful for debugging
	//fmt.Println("NetId:", body.Basic.NetID.Value)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	var person Person
	// Example: Access nested fields
	basic, ok := result["basic"].(map[string]interface{})
	if ok {
		netID, ok := basic["net_id"].(map[string]interface{})
		if ok {
			value, ok := netID["value"].(string)
			if ok {
				fmt.Println("NetID:", value)
				person.netId = value
				person.netIdScoped = value + "@byu.edu"
			}
		}
	}

	if err := json.Unmarshal(body, &person); err != nil {
		return nil, err
	}
	//	person.netId =
	return &person, nil
}

func main() {

	var Token_path = "https://api-sandbox.byu.edu/token" // Sandbox URL
	// var Token_path = "https://api.byu.edu/edu/token" // Production URL

	// Load API credentials from JSON file

	user, pass, err := loadAPICreds("/Users/dwp32/APICreds/in-common.json")
	fmt.Println("User:", user)
	fmt.Println("Pass:", pass)
	if err != nil {
		log.Fatalf("Failed to load API credentials: %v", err)
	}

	// Get the token
	token, err := getToken(Token_path, user, pass)
	if err != nil {
		log.Fatalf("Failed to get token: %v", err)
	}
	fmt.Println("Token:", token)

	// Get the basic user information and load into the Person struct
	person, err := getUserBasicInfo(token, "707501789")
	if err != nil {
		log.Fatalf("Failed to get user info: %v", err)
	}
	fmt.Printf("Person: %+v\n", person)

	// Pull out the basic attributes

	// Set up the HTTP server and routes using Gorilla Mux

	// Create a new Gorilla Mux router
	gRouter := mux.NewRouter()

	// Define the routes and their handlers

	gRouter.HandleFunc("/api/eduPersonAffiliation", getEduPersonAffiliation).Methods("GET")

	gRouter.HandleFunc("/api/name", getName).Methods("GET")
	gRouter.HandleFunc("/api/preferredFirstName", getPreferredFirstName).Methods("GET")
	gRouter.HandleFunc("/api/eduPersonAll", getAll).Methods("GET")
	gRouter.HandleFunc("/api/help", help).Methods("GET")

	// Start the server
	log.Println("Server started on :3000")
	http.ListenAndServe(":3000", gRouter)
}

// initDB initializes the database connection

func getName(w http.ResponseWriter, r *http.Request) {
	consoleMsg := "getName called"
	log.Println(consoleMsg)

	// Get the netId from the query parameters
	netId := r.URL.Query().Get("netId")
	if netId == "" {
		http.Error(w, "Missing netId parameter", http.StatusBadRequest)
		return
	}

	// Set the response header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Return the name as JSON
	/*	response := map[string]string{"name": name}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	*/
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

	// Set the response header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Return the preferredFirstName as JSON
	/*	response := map[string]string{"preferredFirstName": preferredFirstName}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}*/
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

	/*	var affiliations []string

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
	/*	w.Header().Set("Content-Type", "application/json")

		// Return the eduPersonAffiliation as JSON
		response := map[string][]string{"eduPersonAffiliation": affiliations}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}*/
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
	//user := UserInfo{}

	consoleMsg := "getAll called"
	log.Println(consoleMsg)

	/*
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
	*/
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
