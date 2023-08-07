package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// BotEntryHandler logs you in.
func (h *HandlersWithDBStore) BotEntryHandler(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>User Form</title>
</head>
<body>

<div id="userFormDiv" style="display: none;">
    <form id="userForm" action="/submit" method="post">
        <label for="firstName">First Name:</label><br>
        <input type="text" id="firstName" name="firstName" readonly><br>
        <label for="lastName">Last Name:</label><br>
        <input type="text" id="lastName" name="lastName" readonly><br>
        <label for="username">Username:</label><br>
        <input type="text" id="username" name="username" readonly><br>
        <label for="description">Description:</label><br>
        <textarea id="description" name="description" readonly></textarea><br>
        <label for="keywords">Keywords:</label><br>
        <input type="text" id="keywords" name="keywords" readonly><br>
    </form>
</div>

<div id="loginDiv">
    <!-- Login form or message goes here -->
    <p>Please login to see your profile.</p>
</div>

<script>
    // Function to parse URL data
    function parseUrlData() {
        const fragment = window.location.hash.substring(1);
        const params = new URLSearchParams(fragment);
        const userParam = params.get('tgWebAppData');
        const decodedUserParam = decodeURIComponent(userParam);
        const userJson = decodedUserParam.substring(decodedUserParam.indexOf('{'), decodedUserParam.lastIndexOf('}') + 1);
        const user = JSON.parse(decodeURIComponent(userJson));
        return user;
    }

// Function to load the profile page
function loadProfilePage(username) {
    fetch('/api/profile?username=${encodeURIComponent(username)}')
    .then(response => response.json())
    .then(data => {
        // Display the user's profile data
        document.getElementById('firstName').value = data.first_name;
        document.getElementById('lastName').value = data.last_name;
        document.getElementById('username').value = data.username;
        document.getElementById('description').value = data.description;
        document.getElementById('keywords').value = data.keywords;
    });
}
    // Function to show the login page
    function showLoginPage() {
        document.getElementById('userFormDiv').style.display = 'none';
        document.getElementById('loginDiv').style.display = 'block';
    }

    // Function to login the user
    function loginUser(user) {
    	    fetch('/api/login', {
    	method: 'POST',
    	headers: {
    	    'Content-Type': 'application/json'
    	},
    	body: JSON.stringify({
    	    username: user.username
    	})}).then(response => response.json()).then(data => {
	    if (data.success) {
	        // User is authenticated, load the profile page
	        loadProfilePage(user.username);
	    } else {
	        // User is not authenticated, show the login page
	        showLoginPage();
	    }
	});
    }

    // Login the user on page load
    window.onload = function() {
        const user = parseUrlData();
        loginUser(user);
    };
</script>

</body>
</html>
`
	_, err := w.Write([]byte(html))
	if err != nil {
		log.Printf("error 2 %s", err.Error())
	}

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
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	firstName := r.FormValue("firstName")
	lastName := r.FormValue("lastName")
	username := r.FormValue("username")

	fmt.Printf("Received data: firstName = %s, lastName = %s, username = %s\n", firstName, lastName, username)
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
