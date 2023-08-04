package handlers

import (
	"net/http"
)

// Bio returns your bio page.
func (h *HandlersWithDBStore) Bio(w http.ResponseWriter, r *http.Request) {
	html := `
		<!DOCTYPE html>
		<html>
		<head>
		    <title>Your Bio</title>
		    <script src="https://unpkg.com/htmx.org@1.6.1"></script>
    <style>
        body { 
            font-family: Arial, sans-serif; 
            background-color: #000000;
            color: #ffffff;
            margin: 0;
            height: 100vh;
            display: flex;
            justify-content: flex-start;
            align-items: center;
            flex-direction: column;
            padding-left: 20px;
        }
        .content-container {
            text-align: left;
            display: flex;
            flex-direction: column;
            align-items: flex-start;
        }
        .info-block {
            transition: all 0.3s ease;
        }
        .info-block:hover {
            transform: scale(1.05);
        }
        .highlight {
            background-color: #ffffff;
            color: #000000;
            padding: 5px;
            margin: 10px 0;
            transition: all 0.3s ease;
        }
        .highlight:hover {
            background-color: #000000;
            color: #ffffff;
        }
    </style>
		</head>
		<body>
		    <div id="bio" hx-get="/getBio" hx-trigger="load">
		        Loading...
		    </div>
		</body>
		</html>
		`
	_, err := w.Write([]byte(html))
	if err != nil {
		return
	}
}

// GetBio returns your bio.
func (h *HandlersWithDBStore) GetBio(w http.ResponseWriter, r *http.Request) {
	//time.Sleep(time.Second)
	bio := `
		<div>
		    <h1>Emigrant's work.</h1>
		</div>
		`
	_, err := w.Write([]byte(bio))
	if err != nil {
		return
	}
	//http.ServeFile(w, r, "pages/bio.html")
}
