package imageserver

import (
	l4g "code.google.com/p/log4go"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"
)
import _ "net/http/pprof"

const (
	ROOTDIR = "/attachments/paopao"
	PIC404  = "/static/default404.jpg"
	LOGFILE = "imageserver.log"
)

var ReqMap *SafeMap
var log l4g.Logger

func init() {
	initLogger()
	log.Info("CPU num %d", runtime.NumCPU())
	runtime.GOMAXPROCS(runtime.NumCPU())
	ReqMap = NewSafeMap()
}

func Run() {
	go func() {
		log.Info("prof on 6060")
		log.Info(http.ListenAndServe(":6060", nil))
	}()
	go uploadServer()
	thumbServer()
}

func initLogger() {
	log = make(l4g.Logger)
	log.AddFilter("stdout", l4g.DEBUG, l4g.NewConsoleLogWriter())

	flw := l4g.NewFileLogWriter(LOGFILE, false)
	flw.SetRotateDaily(true)
	log.AddFilter("file", l4g.INFO, flw)
}

func thumbServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", imageHandler)

	log.Info("thumb server on 9999")

	s := &http.Server{
		Addr:           ":9999",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        mux,
	}

	log.Error(s.ListenAndServe())
}

func uploadServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", uploadHandler)

	log.Info("upload server on 9998")

	// Schedule cleanup
	s := &http.Server{
		Addr:           ":9998",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        mux,
	}

	log.Error(s.ListenAndServe())
}

func imageHandler(w http.ResponseWriter, r *http.Request) {

	//dynamic
	var req = Req{URI: r.RequestURI}

	if req.Parse() {
		var img, err = getCache(req.hash)

		if len(img) == 0 || err != nil {
		RETRY:
			var q, ok = ReqMap.SetOnce(req.hash)
			if ok {
				img, err = req.AutoResize()

				if err == nil && len(img) > 0 {
					log.Info("gen img done, queue length %d %s", q.length, req.URI)
					ReqMap.DoAndDelete(req.hash, func() {
						setCache(req.hash, img)
						for i := q.length; i > 0; i-- {
							q.ch <- img
						}
						close(q.ch)
					})
				} else {
					ReqMap.DoAndDelete(req.hash, func() {
						close(q.ch)
					})
					log.Error("gen img error %s", req.URI)
				}

			} else {
				img, ok = <-q.ch
				if !ok || len(img) == 0 {
					log.Error("queue receive error <-ch %v", q.ch)
					goto RETRY
				}
			}
		}
		w.Write(img)

	} else {
		w.Header().Set("Location", PIC404)
		w.WriteHeader(http.StatusMovedPermanently)
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "DELETE":
		del(w, r)
	case "POST":
		post(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func post(w http.ResponseWriter, r *http.Request) {
	var req = Req{URI: r.RequestURI}

	if req.Parse() {

		r.ParseMultipartForm(32 << 20)
		file, _, err := r.FormFile("photo")

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			log.Error("upload error %s", []byte(err.Error()))
			return
		}
		defer file.Close()

		err = os.MkdirAll(ROOTDIR+req.path, 0755)
		if err != nil {
			log.Error("create dir error %s", []byte(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		f, err := os.OpenFile(ROOTDIR+req.ori_file, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			log.Error("create file error %s", []byte(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		defer f.Close()
		io.Copy(f, file)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid request"))
		log.Error("parse error %s", req.URI)
	}
}

func del(w http.ResponseWriter, r *http.Request) {

	var req = Req{URI: r.RequestURI}
	if req.Parse() {
		var err = os.Remove(ROOTDIR + req.ori_file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}
