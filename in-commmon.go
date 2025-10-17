/*

in-Common attributes

INPUTS:  netId
OUTPUTS: eduPerson

Example:

/api/eduPerson?byuid=707501789

{
  "netId": "dwp32",
  "netIdScoped": "dwp32@byu.edu",
  "emailAddress": [
    "dwp32@byu.edu",
    ""
  ],
  "primaryAffiliation": "employee",
  "affiliations": [
    "staff",
    "alum",
    "member",
    "affiliate",
    "employee"
  ],
  "scopedAffiliations": [
    "staff@byu.edu",
    "alum@byu.edu",
    "member@byu.edu",
    "affiliate@byu.edu",
    "employee@byu.edu"
  ],
  "name": "David W Palica",
  "preferredFirstName": "Dave"
}


This code is to return the following attributes through an web api call:
Note that BYU does not use 'contractor' or 'library-walk-in' affiliations, but they are included here for completeness.
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
	"slices"
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

	// Define the in-Common affiliation 'triggers'
	faculty, staff, student, alum, member, affiliate, employee := false, false, false, false, false, false, false

	// Define the arrays of group IDs that correspond to each affiliation
	// (pulled from CAS5 /byu-custom/src/main/java/edu/byu/cas/custom/attributes/BYUAttributesSource.java)
	faculty_array := []string{"FULL TIME FACULTY", "CES PERSONNEL", "ROTC", "POST DOC", "VISITING FACULTY", "VISITING SCHOLAR", "PART TIME FACULTY", "AFFILIATE FACULTY"}

	staff_array := []string{"FULL TIME STAFF", "PART TIME STAFF", "Part Time Contract", "PSP", "PURCHASING", "TRAVEL SERVICES", "COOPERATING PROF", "LDS PHILANTHROPIES", "LDS SOC SERV", "CES COMMISSIONERS OFFICE", "EVENING SCHOOL INSTRUCTOR", "INDEPENDENT STUDY INSTRUCTOR", "SALT LAKE CENTER INSTRUCTOR", "CONTINUING ED CONTRACT"}

	student_array := []string{"aerstd", "FULL TIME STUDENT FRESHMAN", "FULL TIME STUDENT SOPHOMORE", "FULL TIME STUDENT JUNIOR", "FULL TIME STUDENT SENIOR", "BGS", "MASTERS PROGRAM", "DOCTORATE PROGRAM", "PART TIME STUDENT", "POST BACCALAUREATE NON DEGREE", "AUDIT", "CONCURRENT ENROLLMENT", "ACADEMIC EXCHANGE", "ELC", "EVENING SCHOOL", "INDEPENDENT STUDY", "SALT LAKE CENTER STUDENT", "VISITING STUDENT"}

	alum_array := []string{"GRADUATED ALUMNI", "FORMER STD--24 COMPLETED HRS"}

	member_array := []string{"FULL TIME FACULTY", "CES PERSONNEL", "ROTC", "POST DOC", "VISITING FACULTY", "VISITING SCHOLAR", "PART TIME FACULTY", "AFFILIATE FACULTY", "aerstd", "FULL TIME STUDENT FRESHMAN", "FULL TIME STUDENT SOPHOMORE", "FULL TIME STUDENT JUNIOR", "FULL TIME STUDENT SENIOR", "BGS", "MASTERS PROGRAM", "DOCTORATE PROGRAM", "PART TIME STUDENT", "POST BACCALAUREATE NON DEGREE", "AUDIT", "CONCURRENT ENROLLMENT", "ACADEMIC EXCHANGE", "ELC", "EVENING SCHOOL", "INDEPENDENT STUDY", "SALT LAKE CENTER STUDENT", "VISITING STUDENT", "FULL TIME STAFF", "PART TIME STAFF", "Part Time Contract", "PSP", "PURCHASING", "TRAVEL SERVICES", "COOPERATING PROF", "LDS PHILANTHROPIES", "LDS SOC SERV", "CES COMMISSIONERS OFFICE", "EVENING SCHOOL INSTRUCTOR", "INDEPENDENT STUDY INSTRUCTOR", "SALT LAKE CENTER INSTRUCTOR", "CONTINUING ED CONTRACT"}

	affiliate_array := []string{"RETIREE", "CRB", "RETIREE SPOUSE", "SURVIVING SPOUSE", "SURVIVING SPOUSE SP", "BYU BENEFITTED", "AFFILIATE FACULTY", "GRADUATED ALUMNI", "EMPLOYEE SPOUSE", "EMPLOYEE DEPENDENT", "SEMINARIES AND INSTITUTES", "PRESIDENTS LEADERSHIP COUNCIL", "BYU WARDS AND STAKES", "SERVICE REPRESENTATIVES", "MTC_Branch", "MTC VOLUNTEER", "WELLS FARGO", "BEEHIVE CLOTHING", "RETIREE DEPENDENT", "STUDENT SPOUSE", "STUDENT DEPENDENT", "FORMER STD--24 COMPLETED HRS", "FULL TIME MISSIONARIES", "FRIENDS OF THE LIBRARY", "DONOR", "NASGuest", "CONTRACT WORKER"}

	employee_array := []string{"FULL TIME FACULTY", "CES PERSONNEL", "ROTC", "POST DOC", "VISITING FACULTY", "VISITING SCHOLAR", "PART TIME FACULTY", "AFFILIATE FACULTY", "FULL TIME STAFF", "PART TIME STAFF", "Part Time Contract", "PSP", "PURCHASING", "TRAVEL SERVICES", "COOPERATING PROF", "LDS PHILANTHROPIES", "LDS SOC SERV", "CES COMMISSIONERS OFFICE", "EVENING SCHOOL INSTRUCTOR", "INDEPENDENT STUDY INSTRUCTOR", "SALT LAKE CENTER INSTRUCTOR", "CONTINUING ED CONTRACT"}

	// Persons V4 API endpoint for group memberships
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

	// Convert the data in to a map for processing (ie, dynamic JSON parsing)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	// Access the "values" array from the JSON structure
	vals, ok := result["values"].([]interface{})
	if !ok {
		log.Fatal("response does not contain 'values' array")
	}

	for _, v := range vals {
		item, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		// This is where we pull out the affiliation information
		// Note that you must do a logical OR so that multiple group_ids can set the same affiliation
		// Otherwise you will only get the last one processed.
		if groupID, ok := item["group_id"].(map[string]interface{}); ok {
			if value, ok := groupID["value"].(string); ok {
				// Check to see which affilations to add.
				faculty = faculty || slices.Contains(faculty_array, value)
				staff = staff || slices.Contains(staff_array, value)
				student = student || slices.Contains(student_array, value)
				alum = alum || slices.Contains(alum_array, value)
				member = member || slices.Contains(member_array, value)
				affiliate = affiliate || slices.Contains(affiliate_array, value)
				employee = employee || slices.Contains(employee_array, value)
				//				person.Affiliations = append(person.Affiliations, value)
				//				person.ScopedAffiliations = append(person.ScopedAffiliations, value+"@byu.edu")
				fmt.Println("group_id:", value)
			}
		}
	}

	if faculty {
		person.Affiliations = append(person.Affiliations, "faculty")
		person.ScopedAffiliations = append(person.ScopedAffiliations, "faculty@byu.edu")
	}
	if staff {
		person.Affiliations = append(person.Affiliations, "staff")
		person.ScopedAffiliations = append(person.ScopedAffiliations, "staff@byu.edu")
	}
	if student {
		person.Affiliations = append(person.Affiliations, "student")
		person.ScopedAffiliations = append(person.ScopedAffiliations, "student@byu.edu")
	}
	if alum {
		person.Affiliations = append(person.Affiliations, "alum")
		person.ScopedAffiliations = append(person.ScopedAffiliations, "alum@byu.edu")
	}
	if member {
		person.Affiliations = append(person.Affiliations, "member")
		person.ScopedAffiliations = append(person.ScopedAffiliations, "member@byu.edu")
	}
	if affiliate {
		person.Affiliations = append(person.Affiliations, "affiliate")
		person.ScopedAffiliations = append(person.ScopedAffiliations, "affiliate@byu.edu")
	}
	if employee {
		person.Affiliations = append(person.Affiliations, "employee")
		person.ScopedAffiliations = append(person.ScopedAffiliations, "employee@byu.edu")
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
