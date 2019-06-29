package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/FJSDS/ws-tcp-relay/logger"
	"go.uber.org/zap"
	"golang.org/x/net/websocket"
)

var tcpAddress string
var binaryMode bool

func copyReader(dst io.Writer, src io.Reader, doneCh chan<- bool) {
	_, _ = io.Copy(dst, src)
	doneCh <- true
}
func copyWriter(dst io.Writer, src io.Reader, doneCh chan<- bool) {
	const maxLen = 64 * 1024
	buf := make([]byte, maxLen)
	headerLen := uint16(4)
	for {
		_, er := src.Read(buf[:headerLen])
		if er != nil {
			break
		}
		totalLen := binary.LittleEndian.Uint16(buf)
		_, er = src.Read(buf[headerLen : totalLen-headerLen])
		if er != nil {
			break
		}

		nw, ew := dst.Write(buf[0:totalLen])
		if ew != nil {
			break
		}
		if int(totalLen) != nw {
			break
		}
	}
	doneCh <- true
}

func relayHandler(ws *websocket.Conn) {
	conn, err := net.DialTimeout("tcp", tcpAddress, time.Second*2)
	if err != nil {
		logger.Info("connect gate error", zap.Error(err))
		return
	}
	if binaryMode {
		ws.PayloadType = websocket.BinaryFrame
	}

	doneCh := make(chan bool)

	go copyReader(conn, ws, doneCh)
	go copyWriter(ws, conn, doneCh)

	<-doneCh
	_ = conn.Close()
	_ = ws.Close()
	<-doneCh
	logger.Info("close")
}

func usage() {
	_, _ = fmt.Fprintf(os.Stderr, "Usage: %s <tcpTargetAddress>\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	var port uint
	var certFile string
	var keyFile string

	flag.UintVar(&port, "p", 8080, "The port to listen on")
	flag.UintVar(&port, "port", 8080, "The port to listen on")
	flag.StringVar(&certFile, "tlscert", "", "TLS cert file path")
	flag.StringVar(&keyFile, "tlskey", "", "TLS key file path")
	flag.BoolVar(&binaryMode, "b", false, "Use binary frames instead of text frames")
	flag.BoolVar(&binaryMode, "binary", false, "Use binary frames instead of text frames")
	flag.Usage = usage
	flag.Parse()

	tcpAddress = flag.Arg(0)
	if tcpAddress == "" {
		_, _ = fmt.Fprintln(os.Stderr, "No address specified")
		os.Exit(1)
	}
	l, _ := logger.New("ws-tcp-relay", ".", zap.DebugLevel)
	logger.SetDefaultLog(l)
	portString := fmt.Sprintf(":%d", port)

	logger.Info("Listening addr", zap.String("addr", portString))
	http.Handle("/ws", websocket.Handler(relayHandler))
	http.Handle("/", http.HandlerFunc(home))
	var err error
	if certFile != "" && keyFile != "" {
		err = http.ListenAndServeTLS(portString, certFile, keyFile, nil)
	} else {
		err = http.ListenAndServe(portString, nil)
	}
	if err != nil {
		logger.Error("listen error", zap.Error(err))
	}
}
func home(w http.ResponseWriter, r *http.Request) {
	_ = homeTemplate.Execute(w, "ws://"+r.Host+"/ws")
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
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
		ws.binaryType = 'arraybuffer';
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
<p>Click "Open" to create a connection to the server, 
"Send" to send a message to the server and "Close" to close the connection. 
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output"></div>
</td></tr></table>
</body>
</html>
`))
