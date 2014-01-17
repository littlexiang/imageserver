package imageserver

import (
	"io"
	"net/http"
	"os"
	"runtime"
	"time"
)
import _ "net/http/pprof"

var ReqMap *SafeMap

func init() {
	initConf()
	initLogger()
	initPool()
	ReqMap = NewSafeMap()

	log.Info("CPU num %d", runtime.NumCPU()-1)
	runtime.GOMAXPROCS(runtime.NumCPU() - 1)
}

func Run() {
	go statsServer()
	go uploadServer()
	thumbServer()
}

func statsServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", statsHandler)

	log.Info("stats server on 9997")

	s := &http.Server{
		Addr:           ":9997",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        mux,
	}

	log.Error(s.ListenAndServe())
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
	var req = Req{URI: r.RequestURI}
	S.IncGet()

	start := time.Now()
	defer func() {
		end := time.Now()
		log.Info("fetch %s: %f s", req.URI, end.Sub(start).Seconds())
	}()

	if req.Parse() {

		var img, err = getCache(req.hash)

		if len(img) == 0 || err != nil {
		RETRY:
			var q, ok = ReqMap.SetOnce(req.hash)
			if ok {
				img, err = req.AutoResize()

				if err == nil && len(img) > 0 {
					log.Warn("regen img, queue length %d %s", q.length, req.URI)
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
					log.Error("gen img error %s %v", req.URI, err)
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
		w.Header().Set("Location", C.PIC404)
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

		err = os.MkdirAll(C.ROOTDIR+req.path, 0755)
		if err != nil {
			log.Error("create dir error %s", []byte(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		f, err := os.OpenFile(C.ROOTDIR+req.ori_file, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			log.Error("create file error %s", []byte(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		defer f.Close()
		io.Copy(f, file)
		S.IncPost()
		log.Info("upload %s", req.URI)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid request"))
		log.Error("parse error %s", req.URI)
	}
}

func del(w http.ResponseWriter, r *http.Request) {
	var req = Req{URI: r.RequestURI}
	if req.Parse() {
		var err = os.Remove(C.ROOTDIR + req.ori_file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		delCache(req.hash)
		S.IncDelete()
		log.Info("delete %s", req.URI)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}
