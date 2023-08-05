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
			<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/5.15.4/css/all.min.css">
			<title>Bio</title>
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
				.addTextButton {
				    padding: 10px 20px; /* size */
				    background-color: #008CBA; /* color */
				    border: none; /* remove default border */
				    color: black; /* text color */
				    text-align: center; /* align text */
				    text-decoration: none; /* remove underline */
				    display: inline-block;
				    font-size: 16px; /* text size */
				    margin: 4px 2px;
				    transition-duration: 0.4s; /* transition effect */
				    cursor: pointer; /* change cursor style on hover */
				    border-radius: 4px; /* rounded corners */
				}
    		</style>
		</head>
		<body>
			<div id="bio" hx-get="/getBio" hx-trigger="load">
			        Loading...
			</div>
			<button id="addTextButton" class="addTextButton">Copy</button>
    		<div id="extraText"><h1></h1></div>
    		<script>
    		    document.getElementById("addTextButton").addEventListener("click", function() {
    		        var newText = document.createElement("p");
    		        newText.textContent = "...is to stay kind, humble and keep learning.";
    		        document.getElementById("extraText").appendChild(newText);
    		        this.style.display = "none"; // hides the button
    		    });
    		</script>

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
		    <h1>Emigrant's work...</h1>
		</div>
		`
	_, err := w.Write([]byte(bio))
	if err != nil {
		return
	}
	//http.ServeFile(w, r, "pages/bio.html")
}
