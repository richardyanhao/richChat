package main

import (
	"net/http"
	"chatroom"
	"log"
	"html/template"
)

func chat(w http.ResponseWriter, r *http.Request)  {
	log.Println("before upgrade")
	c, err := chatroom.Upgrade(w, r)
	log.Println("after upgrade")
	if err != nil {
		log.Println("upgrade error", err)
		return
	}
	defer c.Close()
	for {
		message, err := c.ReadData()
		if err != nil {
			log.Println("read:", err)
			break
		}
		for _, v := range(chatroom.ConnMap) {
			if v != c {
				v.WriteData(message)
			}
		}
		log.Printf("resv:%s", message)
	}
}

func main()  {
	log.SetFlags(0)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		homeTemplate.Execute(w, "ws://"+r.Host+"/chatroom")
	})
	http.HandleFunc("/chatroom", chat)
	http.ListenAndServe("127.0.0.1:2333", nil)
}

// index 页面内容
var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<head>
<meta charset="utf-8">
<script>
window.addEventListener("load", function(evt) {

    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;

    var print = function(message) {
        var d = document.createElement("div");
        d.innerHTML = message;
        output.appendChild(d);
    };

    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };

    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);
        ws.send(input.value);
        return false;
    };

    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };

});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>
点击 "Open" 开始一个新的WebSocket链接,
“Send" 将内容发送到服务器，
"Close" 将关闭链接。
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="hello I am Richard!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output"></div>
</td></tr></table>
</body>
</html>
`))
