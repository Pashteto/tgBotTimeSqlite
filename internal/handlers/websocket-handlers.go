package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func (h *HandlersWithDBStore) EchoWS(w http.ResponseWriter, r *http.Request) {
	log.Println("got the EchoWS request, Time: ", time.Now().String())
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	go func() {
		for {
			messageType, message, err := conn.ReadMessage()
			log.Println("got message: ", message, "; Time: ", time.Now().String())
			if err != nil {
				log.Println(err)
				return
			}
			dfvlnd := string(message)
			dfvlnd = strings.ReplaceAll(dfvlnd, " ", "")
			log.Println("sent back: ", dfvlnd, "; Time: ", time.Now().String())
			err = conn.WriteMessage(messageType, []byte(dfvlnd))
			if err != nil {
				log.Println(err)
				return
			}
		}
	}()
}

func (h *HandlersWithDBStore) GetTestTime(w http.ResponseWriter, r *http.Request) {
	log.Println("got the GetTestTime request, Time: ", time.Now().String())
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	addressPort := h.Conf.WebSocketEnd
	w.Write([]byte(`
		<!DOCTYPE html>
<html>
<body>
    <input id="input" type="text" onkeydown="if (event.keyCode == 13) send();">
    <button onclick="send()">Send</button>
    <ul id="messages"></ul>
    <style>
        .client {
            text-align: right;
            color: blue;
        }

        .server {
            text-align: left;
            color: green;
        }
    </style>
    <script>
		
` +
		fmt.Sprintf(`var ws = new WebSocket('%s');`, addressPort) + `
        ws.onmessage = function(event) {
            var messages = document.getElementById('messages');
            var message = document.createElement('li');
            var now = new Date();
            var time = ('0' + now.getHours()).slice(-2) + ':' + 
                       ('0' + now.getMinutes()).slice(-2) + ':' + 
                       ('0' + now.getSeconds()).slice(-2);
            message.innerText = 'Server says: ' + event.data.split("").reverse().join("") + ' : ' + time;
            message.className = 'server';
            messages.appendChild(message);
        };

        function send() {
            var input = document.getElementById('input');
            var messages = document.getElementById('messages');
            var message = document.createElement('li');
            var now = new Date();
            var time = ('0' + now.getHours()).slice(-2) + ':' + 
                       ('0' + now.getMinutes()).slice(-2) + ':' + 
                       ('0' + now.getSeconds()).slice(-2);
            message.innerText = 'You: ' + input.value + ' : ' + time;
            message.className = 'client';
            messages.appendChild(message);
            ws.send(input.value);
            input.value = '';
        }

        ws.onopen = function() {
            console.log('Connection opened!');
        };

        ws.onclose = function() {
            console.log('Connection closed');
        };

        ws.onerror = function(err) {
            console.log('Error occurred: ', err);
        };
    </script>
</body>
</html>`))
}
