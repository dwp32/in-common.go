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
	"time"

	"github.com/gorilla/mux"
)

var Token_path = "https://api-sandbox.byu.edu/token" // Sandbox URL
// var Token_path = "https://api.byu.edu/edu/token" // Production URL

var Token string

type APICreds struct {
	APIUser string `json:"API_USER"`
	APIPass string `json:"API_PASS"`
}

type Person struct {
	NetID              string   `json:"netId"`
	NetIDScoped        string   `json:"netIdScoped"`
	EmailAddress       []string `json:"emailAddress"`
	PrimaryAffiliation string   `json:"primaryAffiliation"`
	Affiliations       []string `json:"affiliations"`
	ScopedAffiliations []string `json:"scopedAffiliations"`
	Name               string   `json:"name"`
	PreferredFirstName string   `json:"preferredFirstName"`
}

// This is the structure that contains all of the information. All or parts may be returned, depend

func startTimer(duration time.Duration) {
	timer := time.NewTimer(duration)

	// Load API credentials from JSON file
	user, pass, err := loadAPICreds("/Users/dwp32/APICreds/in-common.json")
	fmt.Println("User:", user)
	fmt.Println("Pass:", pass)
	if err != nil {
		log.Fatalf("Failed to load API credentials: %v", err)
	}
	go func() {
		for {
			<-timer.C
			fmt.Println("Token expired! Renewing...")
			fmt.Println(time.Now().Format("2006-01-02 15:04:05"))
			// Reset the timer to start again
			timer.Reset(duration)
			err = setGlobalToken(Token_path, user, pass)
			fmt.Println("Token is now: ", Token)
		}
	}()
}

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

func setGlobalToken(endpoint, user, pass string) error {

	var ok bool
	// Build form body
	form := url.Values{}
	form.Add("grant_type", "client_credentials")

	// Create POST request with the form body
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(user, pass)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read and decode the response
	body, _ := io.ReadAll(resp.Body)
	//fmt.Println("Raw response:", string(body)) // helpful for debugging

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get token: %s", resp.Status)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	// Depending on the API, the key might be "access_token" instead of "token"
	Token, ok = result["access_token"].(string)
	if !ok {
		// Try alternate key just in case
		Token, ok = result["token"].(string)
		if !ok {
			return fmt.Errorf("Token not found in response")
		}
	}

	//	fmt.Println("Token:", Token)
	return nil
}

// Group membership information
func getUserGroupInfo(person *Person, netId string) error {

	endpoint := fmt.Sprintf("https://api-sandbox.byu.edu/byuapi/persons/v4/%s/group_memberships", netId)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Response Status:", resp.Status)
		return fmt.Errorf("failed to get group membership info: %s", resp.Status)
	}

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Raw group info response:", string(body)) // helpful for debugging

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	//vals, ok := result["values"].(map[string]interface{})

	vals, ok := result["values"].([]interface{})
	if !ok {
		log.Fatal("response does not contain 'values' array")
	}

	for _, v := range vals {
		item, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		if groupID, ok := item["group_id"].(map[string]interface{}); ok {
			if value, ok := groupID["value"].(string); ok {
				person.Affiliations = append(person.Affiliations, value)
				person.ScopedAffiliations = append(person.ScopedAffiliations, value+"@byu.edu")
				fmt.Println("group_id:", value)
			}
		}
	}

	return nil
}

// Basic user information
func getUserBasicInfo(person *Person, netId string) error {

	// Example endpoint, replace with actual
	endpoint := fmt.Sprintf("https://api-sandbox.byu.edu/byuapi/persons/v4/%s", netId)

	//var person Person
	var first, middle, last string
	//var email string

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Response Status:", resp.Status)
		return fmt.Errorf("failed to get user info: %s", resp.Status)
	}

	body, _ := io.ReadAll(resp.Body)
	//fmt.Println("Raw user info response:", string(body)) // helpful for debugging
	//fmt.Println("NetId:", body.Basic.NetID.Value)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	// Access nested fields from JSON structure

	// NetID is located at result["basic"]["net_id"]["value"]
	basic, ok := result["basic"].(map[string]interface{})
	if ok {
		netID, ok := basic["net_id"].(map[string]interface{})
		if ok {
			value, ok := netID["value"].(string)
			if ok {
				fmt.Println("Processing NetID:", value)
				person.NetID = value
				person.NetIDScoped = value + "@byu.edu"
			}
		}

		// Extract Preferred First Name
		if pfn, ok := basic["preferred_first_name"].(map[string]interface{}); ok {
			if value, ok := pfn["value"].(string); ok {
				fmt.Println("Preferred First Name:", value)
				person.PreferredFirstName = value
			}
		}

		// Extract Full Name

		// Full first name
		if firstname, ok := basic["first_name"].(map[string]interface{}); ok {
			if value, ok := firstname["value"].(string); ok {
				fmt.Println("First Name:", value)
				first = value
			}
		}

		// Full middle name
		if middlename, ok := basic["middle_name"].(map[string]interface{}); ok {
			if value, ok := middlename["value"].(string); ok {
				fmt.Println("Middle Name:", value)
				middle = value
			}
		}

		// Full last name
		if lastname, ok := basic["surname"].(map[string]interface{}); ok {
			if value, ok := lastname["value"].(string); ok {
				fmt.Println("Last Name:", value)
				last = value
			}
		}

		person.Name = first + " " + middle + " " + last

		// There are three email addresses:
		// 	personal_email_address,
		//  student_email_address,
		//  byu_intenral_email
		if email, ok := basic["byu_internal_email"].(map[string]interface{}); ok {
			if value, ok := email["value"].(string); ok {
				person.EmailAddress = append(person.EmailAddress, value)
			}
		}

		if email, ok := basic["student_email_address"].(map[string]interface{}); ok {
			if value, ok := email["value"].(string); ok {
				person.EmailAddress = append(person.EmailAddress, value)
			}
		}
		if email, ok := basic["personal_email"].(map[string]interface{}); ok {
			if value, ok := email["value"].(string); ok {
				person.EmailAddress = append(person.EmailAddress, value)
			}
		}
	}

	if err := json.Unmarshal(body, &person); err != nil {
		return err
	}
	//	person.netId =
	fmt.Println(person)
	return nil
}

func main() {

	// Load API credentials from JSON file

	user, pass, err := loadAPICreds("/Users/dwp32/APICreds/in-common.json")
	fmt.Println("User:", user)
	fmt.Println("Pass:", pass)
	if err != nil {
		log.Fatalf("Failed to load API credentials: %v", err)
	}

	// Set the initial global token
	err = setGlobalToken(Token_path, user, pass)

	// Renew the global token as needed. Here we set it to renew every 10 seconds for testing purposes. In production, set it to 3500 seconds (just before the 1 hour expiration).
	// The timer runs in a separate goroutine.
	startTimer(3500 * time.Second)

	if err != nil {
		log.Fatalf("Failed to get token: %v", err)
	}
	fmt.Println("Global Token:", Token)

	// Set up the HTTP server and routes using Gorilla Mux

	// Create a new Gorilla Mux router
	gRouter := mux.NewRouter()

	// Define the routes and their handlers

	gRouter.HandleFunc("/api/eduPerson", eduPerson).Methods("GET")
	gRouter.HandleFunc("/api/help", help).Methods("GET")

	// Start the server
	log.Println("Server started on :3000")
	http.ListenAndServe(":3000", gRouter)
}

// initDB initializes the database connection

func eduPerson(w http.ResponseWriter, r *http.Request) {
	person := Person{}

	// Get the byuId from the query parameters
	byuId := r.URL.Query().Get("byuId")
	if byuId == "" {
		http.Error(w, "Missing byuId parameter", http.StatusBadRequest)
		return
	}

	// Fetch user info using the provided byuId
	err := getUserBasicInfo(&person, byuId)
	if err != nil {
		log.Printf("Failed to get user info: %v", err)
		http.Error(w, "Failed to retrieve user info", http.StatusInternalServerError)
		return
	}

	// Fetch user group memberships using the provided byuId
	err = getUserGroupInfo(&person, byuId)
	if err != nil {
		log.Printf("Failed to get group info: %v", err)
		http.Error(w, "Failed to retrieve group info", http.StatusInternalServerError)
		return
	}

	// Set the response header to application/json
	w.Header().Set("Content-Type", "application/json")
	//var persont = `{"netId": "jdoe"}`
	// Encode struct to JSON and write the response
	if err := json.NewEncoder(w).Encode(person); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	// Optional: log the person struct to console
	fmt.Printf("Person: %+v\n", person)
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
