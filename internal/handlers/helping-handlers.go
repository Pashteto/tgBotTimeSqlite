package handlers

import "net/http"

// GetNikitaReq listens to bot.
func (h *HandlersWithDBStore) GetNikitaReq(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
    <title>Information Page</title>
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
    <div class="content-container">
        <div class="info-block">
            <h2>Сбор на BOG:</h2>
            <p class="highlight">GE26BG0000000533615481</p>
            <p>Nikita Klimov</p>
        </div>
        <div class="info-block">
            <h2>На Тинькоф:</h2>
            <p class="highlight">+79950905198</p>
        </div>
    </div>
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
    <title>Information Page</title>
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
    <div class="content-container">
        <div class="info-block">
            <h2>TBC Bank:</h2>
            <p class="highlight">GE54TB7203645064300043</p>
            <p>Elena Mavromatis</p>
        </div>
        <div class="info-block">
            <h2>Swift code:</h2>
            <p class="highlight">TBCBGE22</p>
        </div>
    </div>
</body>
</html>`))
}
