package handlers

import "net/http"

// GetNikitaReq listens to bot.
func (h *HandlersWithDBStore) GetNikitaReq(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
    <title>Bank details</title>
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
		.copyButton {
		    //padding: 10px 20px; /* size */
		    //background-color: #008CBA; /* color */
		    //border: none; /* remove default border */
		    //color: white; /* text color */
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
    <div class="content-container">
        <div class="info-block">
            <h2>Сбор на BOG:</h2>
			<p class="highlight"><span id="itemBOGN">GE26BG0000000533615481</span>  <button class="copyButton" data-copy-target="#itemBOGN">Copy</button></p>
			<p>Nikita Klimov</p>
        </div>
        <div class="info-block">
            <h2>На Тинькоф:</h2>
            <p class="highlight"><span id="itemPhoneCopy">+79950905198</span>  <button class="copyButton" data-copy-target="#itemPhoneCopy">Copy</button></p>
        </div>
    </div>
	<script>
	    document.querySelectorAll(".copyButton").forEach(function(button) {
	        button.addEventListener("click", async function() {
	            var copyTargetId = this.getAttribute("data-copy-target");
	            var copyTarget = document.querySelector(copyTargetId);
	            if(copyTarget) {
	                try {
	                    await navigator.clipboard.writeText(copyTarget.textContent);
	                    console.log('Copying to clipboard was successful!');
	                } catch (err) {
	                    console.error('Failed to copy text: ', err);
	                }
	            }
	        });
	    });
	</script>
</body>
</html>`))
}

// ==========================================================================================================================================================

// GetElenaReq listens to bot.
func (h *HandlersWithDBStore) GetElenaReq(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
    <title>Bank details</title>
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
		.copyButton {
		    //padding: 10px 20px; /* size */
		    //background-color: #008CBA; /* color */
		    //border: none; /* remove default border */
		    //color: white; /* text color */
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
    <div class="content-container">
        <div class="info-block">
            <h2>TBC Bank:</h2>
			<p class="highlight"><span id="itemBOGE">GE54TB7203645064300043</span>  <button class="copyButton" data-copy-target="#itemBOGE">Copy</button></p>
            <p>Elena Mavromatis</p>
        </div>
        <div class="info-block">
            <h2>Swift code:</h2>
			<p class="highlight"><span id="itemSwift">TBCBGE22</span>  <button class="copyButton" data-copy-target="#itemSwift">Copy</button></p>
        </div>
    </div>
	<script>
	    document.querySelectorAll(".copyButton").forEach(function(button) {
	        button.addEventListener("click", async function() {
	            var copyTargetId = this.getAttribute("data-copy-target");
	            var copyTarget = document.querySelector(copyTargetId);
	            if(copyTarget) {
	                try {
	                    await navigator.clipboard.writeText(copyTarget.textContent);
	                    console.log('Copying to clipboard was successful!');
	                } catch (err) {
	                    console.error('Failed to copy text: ', err);
	                }
	            }
	        });
	    });
	</script>
</body>
</html>`))
}
