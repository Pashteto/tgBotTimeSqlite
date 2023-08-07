package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/gorilla/mux"

	"github.com/Pashteto/tgBotTimeSqlite/config"
	filedb "github.com/Pashteto/tgBotTimeSqlite/filed_history"
	"github.com/Pashteto/tgBotTimeSqlite/internal/handlers"
	"github.com/Pashteto/tgBotTimeSqlite/internal/repo"
)

var ctx = context.Background()

func main() {
	var conf config.Config

	ServAddrPtr := flag.String("a", ":8080", "SERVER_ADDRESS")
	BaseURLPtr := flag.String("b", "http://localhost:8080", "BASE_URL")
	FStorPathPtr := flag.String("f", "../possible_data", "FILE_STORAGE_PATH")
	flag.Parse()

	log.Println("Flags input:\nSERVER_ADDRESS,\tBASE_URL,\tFILE_STORAGE_PATH:\t", *ServAddrPtr, ",", *BaseURLPtr, ",", *FStorPathPtr)
	err := env.Parse(&conf)
	if err != nil {
		log.Fatalf("Unable to Parse env:\t%v", err)
	}
	log.Printf("Config:\t%+v", conf)

	changed, err := conf.UpdateByFlags(ServAddrPtr, BaseURLPtr, FStorPathPtr)
	if changed {
		log.Printf("Config updated:\t%+v\n", conf)
	}
	if err != nil {
		log.Printf("Flags input error:\t%v\n", err)
	}

	log.Println("REDIS_HOST:\t", os.Getenv("REDIS_HOST"))
	log.Println("USER:\t", os.Getenv("USER"))

	err = filedb.CreateDirFileDBExists(conf)
	if err != nil {
		log.Fatalf("file exited;\nerr:\t%v", err)
	}

	// Open the database file, creating it if necessary
	sqliteDB, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal(err)
	}
	defer sqliteDB.Close()

	sqliteRepo := repo.NewRepo(sqliteDB)
	err = sqliteRepo.CreateTableIfNotExists(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fileServ := http.FileServer(http.Dir("./templates/**/*"))

	sshand := handlers.NewHandlersWithDBStore(&conf, sqliteRepo, &fileServ)

	r := mux.NewRouter()

	r.HandleFunc("/stop_listen", sshand.StopListenBot).Methods("POST")  //routing post
	r.HandleFunc("/listen", sshand.ListenBot).Methods("POST")           //routing post
	r.HandleFunc("/helping-nikita", sshand.GetNikitaReq).Methods("GET") //routing get
	r.HandleFunc("/helping-elena", sshand.GetElenaReq).Methods("GET")   //routing get
	r.HandleFunc("/echo", sshand.EchoWS).Methods("GET")                 //routing post
	r.HandleFunc("/bot",
		func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(1 * time.Microsecond)
			log.Printf("header %+v", r.Header)
			log.Printf("body %+v", r.Body)
			var data []byte
			read, err := r.Body.Read(data)
			if err != nil {
				log.Printf("error %s", err.Error())
			} else {
				log.Printf("read %d: %s", read, string(data))
			}
			log.Printf("body %+v", r.Body)
			log.Printf("cookies %+v", r.Cookies())
			log.Printf("URL %+v", r.URL)
			log.Printf("Form %+v", r.Form)
			log.Printf("RequestURI %+v", r.RequestURI)
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
			<button id="addTextButton" class="addTextButton">...?</button>
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
			_, err = w.Write([]byte(html))
			if err != nil {
				log.Printf("error 2 %s", err.Error())
			}
			//sshand.GetNikitaReq(w, r)
			//log.Println("bot-login, redirecting to https://www.google.com/ or http://localhost:8181 + :", r.RequestURI)
			//link := "http://localhost:8181" + strings.TrimPrefix(r.RequestURI, "/bot")
			//http.Redirect(w, r, "https://www.google.com/", http.StatusMovedPermanently)
		})
	r.HandleFunc("/get_test_time", sshand.GetTestTime).Methods("GET") //routing post
	r.HandleFunc("/getBio", sshand.GetBio).Methods("GET")             //routing post
	r.HandleFunc("/", sshand.Bio).Methods("GET")                      //routing post

	http.Handle("/", r)

	// конструируем свой сервер
	server := &http.Server{
		Addr: conf.ServAddr,
		//TLSConfig: tlsConfig,
	}

	sigint := make(chan os.Signal)
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	go func() {
		sig := <-sigint // Blocks here until interrupted
		log.Println(sig, "\t<<<===\t signal received. Shutdown process initiated.")
		err := server.Shutdown(ctx)
		if err != nil {
			log.Println(err)
			return
		}
	}()

	if conf.LocalDebug {
		httpMux := http.NewServeMux()
		httpMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "http://localhost:8080"+r.RequestURI, http.StatusMovedPermanently)
		})
		go func() {
			log.Fatal(http.ListenAndServe(":80", httpMux))
		}()

		err = server.ListenAndServe()
		if err != nil {
			log.Println(err.Error())
			return
		}
	} else {
		httpMux := http.NewServeMux()
		httpMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "https://localhost:443"+r.RequestURI, http.StatusMovedPermanently)
		})
		go func() {
			log.Fatal(http.ListenAndServe(":80", httpMux))
		}()
		server.Addr = ":443"
		err = server.ListenAndServeTLS("/etc/letsencrypt/live/pashteto.com/fullchain.pem", "/etc/letsencrypt/live/pashteto.com/privkey.pem")
		if err != nil {
			log.Println(err.Error())
			return
		}
	}
}
