package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// BotEntryHandler logs you in.
func (h *HandlersWithDBStore) BotEntryHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/html/index.html")
	//_, err := w.Write([]byte(html))
	//if err != nil {
	//	log.Printf("error 2 %s", err.Error())
	//}
}

// BotLogin BotLogin.
func (h *HandlersWithDBStore) BotLogin(w http.ResponseWriter, r *http.Request) {
	log.Println("BotLogin r", r)
	// Decode the request body
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Println("BotLogin", req)

	// Check if the user exists in our "database"
	dbMu.RLock()
	_, ok := db[req.Username]
	dbMu.RUnlock()
	log.Println("BotLogin _, ok := db[req.Username]: ", ok)

	// Respond with the result
	json.NewEncoder(w).Encode(map[string]bool{"success": ok})
}

func (h *HandlersWithDBStore) BotProfile(w http.ResponseWriter, r *http.Request) {
	log.Println("BotProfile", r)
	// Get the username from the query parameters
	username := r.URL.Query().Get("username")

	// Look up the user in our "database"
	dbMu.RLock()
	user, ok := db[username]
	dbMu.RUnlock()

	// If the user doesn't exist, respond with an error
	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Respond with the user's profile data
	json.NewEncoder(w).Encode(user)
}

// BotSubmitHandler fetches your fragments.
func (h *HandlersWithDBStore) BotSubmitHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Parse and validate the data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	firstName := r.FormValue("firstName")
	lastName := r.FormValue("lastName")
	username := r.FormValue("username")
	description := r.FormValue("description")
	keywords := r.FormValue("keywords")

	fmt.Printf("Received data: firstName = %s, lastName = %s, username = %s\n", firstName, lastName, username)

	// 2. Save the data to the database map
	newUser := User{
		FirstName:   firstName,
		LastName:    lastName,
		Username:    username,
		Description: description,
		Keywords:    keywords,
	}

	dbMu.Lock() // Lock the mutex for writing
	db[username] = newUser
	dbMu.Unlock() // Unlock the mutex after writing

	// 3. Redirect to the base page
	http.Redirect(w, r, "/bo", http.StatusSeeOther) // StatusSeeOther (303) is used to redirect after a POST
}

// Rotate Rotate.
func (h *HandlersWithDBStore) Rotate(w http.ResponseWriter, r *http.Request) {
	defer func(start time.Time) {
		log.Println("Rotate redirected in: ", time.Since(start).Microseconds(), " Micro seconds")
	}(time.Now())

	targetURL := "http://129.146.183.89:8904" + strings.TrimPrefix(r.RequestURI, "/bo")

	http.Redirect(w, r, targetURL, http.StatusSeeOther) // StatusSeeOther (303) is used to redirect after a POST
}

/*
// ParseFragmHandler fetches your fragments.
func (h *HandlersWithDBStore) ParseFragmHandler(w http.ResponseWriter, r *http.Request) {
	var dataAA struct {
		Fragment string `json:"fragment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&dataAA); err != nil {
		log.Printf("Could not decode body: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	// At this point, data.Fragment contains the URL fragment.
	log.Printf("Received fragment: %s", dataAA.Fragment)

	params, userData, err := parseFragment(dataAA.Fragment)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Parameters: %#v\n", params)
	fmt.Printf("User Data: %#v\n", userData)
}
/*
type UserData struct {
	ID              int    `json:"id"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	Username        string `json:"username"`
	LanguageCode    string `json:"language_code"`
	AllowsWriteToPm bool   `json:"allows_write_to_pm"`
}

*/
/*func parseFragment(fragment string) (map[string]string, UserData, error) {
	params := make(map[string]string)
	var userData UserData

	// Remove the leading "#"
	if len(fragment) > 0 && fragment[0] == '#' {
		fragment = fragment[1:]
	}

	// URL decode the fragment
	decoded, err := url.QueryUnescape(fragment)
	if err != nil {
		return nil, UserData{}, err
	}

	// Split the fragment into parameters
	for _, param := range strings.Split(decoded, "&") {
		parts := strings.SplitN(param, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		if key == "user" {
			// The user data is provided as a URL-encoded JSON string
			decodedUser, err := url.QueryUnescape(value)
			if err != nil {
				return nil, UserData{}, err
			}

			// Parse the JSON string
			err = json.Unmarshal([]byte(decodedUser), &userData)
			if err != nil {
				return nil, UserData{}, err
			}
		} else {
			params[key] = value
		}
	}

	return params, userData, nil
}
*/
type User struct {
	Username    string `json:"username"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Description string `json:"description"`
	Keywords    string `json:"keywords"`
}

type LoginRequest struct {
	Username string `json:"username"`
}

var (
	db   = map[string]User{} // This will be our simple "database"
	dbMu sync.RWMutex        // This mutex will ensure our "database" is safe to use concurrently
)
