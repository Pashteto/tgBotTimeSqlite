package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
			time.Sleep(1 * time.Second)
			log.Println("bot-login, redirecting to http://localhost:8181 + :", r.RequestURI)
			link := "http://localhost:8181" + strings.TrimPrefix(r.RequestURI, "/bot")
			http.Redirect(w, r, link, http.StatusMovedPermanently)
		})
	r.HandleFunc("/get_test_time", sshand.GetTestTime).Methods("GET") //routing post
	r.HandleFunc("/", sshand.Bio).Methods("GET")                      //routing post
	r.HandleFunc("/getBio", sshand.GetBio).Methods("GET")             //routing post

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
