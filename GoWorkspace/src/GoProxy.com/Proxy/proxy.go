package main

import (
	"bytes"
	"flag"
	"hash/crc32"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

type MyResponseWriter struct {
	http.ResponseWriter
	buf *bytes.Buffer
}

func (mrw *MyResponseWriter) Write(p []byte) (int, error) {
	return mrw.buf.Write(p)
}
func handleHTTPS(w http.ResponseWriter, r *http.Request) {
	log.Println("URL: ", r.RequestURI, " METHOD: ", r.Method)
	http.Error(w, "https not supported", http.StatusBadRequest)
}
func handleHTTP(w http.ResponseWriter, req *http.Request) {
	log.Println("URL: ", req.RequestURI, " METHOD: ", req.Method)

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	log.Println(req.RemoteAddr, "", resp.Status)

	defer resp.Body.Close()

	bodyBytes, bodyErr := ioutil.ReadAll(resp.Body)
	if bodyErr != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	crc := crc32.ChecksumIEEE(bodyBytes)
	log.Println("CRC32", crc)

	resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)

}
func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
func main() {
	log.Println("Starting ...")

	//read from command line args
	var listenPort int
	flag.IntVar(&listenPort, "port", 9087, "Listen Port")
	flag.Parse()
	log.Println("listening on port: ", listenPort)

	server := &http.Server{
		Addr: ":" + strconv.Itoa(listenPort),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println("protocol: ", r.Host)
			if r.Method == http.MethodConnect {
				handleHTTPS(w, r)
			} else {
				handleHTTP(w, r)
			}

		}),
	}

	log.Fatal(server.ListenAndServe())
}
