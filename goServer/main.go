package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"encoding/json"
	//"strconv"
)
type apiConfig struct {
	fileserverHits atomic.Int32
}
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {

	return http.HandlerFunc( func(w http.ResponseWriter, req *http.Request){
			cfg.fileserverHits.Add(1)
			cfg.fileserverHits.Load()
			next.ServeHTTP(w,req)
	})
}
func (cfg *apiConfig) getHitsString() string {
	html := "<html>"+
				"<body>" +
				"<h1>Welcome, Chirpy Admin</h1>"+
				"<p>Chirpy has been visited %d times!</p>" +
				"</body>"+
			"</html>"
	return fmt.Sprintf(html,cfg.fileserverHits.Load())
}
func (cfg *apiConfig) reset(next http.Handler) http.Handler {
	cfg.fileserverHits.Store(0)
	return next
}
func main(){
	const port = "8080"
	mux := http.NewServeMux()
	
	var apiCfg apiConfig
	fileServer :=  http.FileServer(http.Dir("."))
	handler := http.StripPrefix("/app/", fileServer)
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))
	mux.HandleFunc("POST /admin/reset", func(w http.ResponseWriter, req *http.Request){
		w.Header().Add("Content-Type","text/plain; charset=utf-8" )
		w.WriteHeader(200)
		w.Write([]byte("Reset"))
	})
	mux.HandleFunc("GET /admin/metrics", func(w http.ResponseWriter, req *http.Request){
		w.Header().Add("Content-Type","text/html; charset=utf-8" )
		w.WriteHeader(200)
		w.Write([]byte(apiCfg.getHitsString()))
	})
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, req *http.Request){
		w.Header().Add("Content-Type","text/plain; charset=utf-8" )
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("POST /api/validate_chirp", func(w http.ResponseWriter, req *http.Request){
		type parameters struct {
			// these tags indicate how the keys in the JSON should be mapped to the struct fields
			// the struct fields must be exported (start with a capital letter) if you want them parsed
			Body string `json:"body"`
		}
		type returnVals struct {
			// the key will be the name of struct field unless you give it an explicit JSON tag
			Valid bool `json:"valid"`
			Error string `json:"error"`
		}
		
		decoder := json.NewDecoder(req.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			responseError := returnVals{
				Error: "Something went wrong",
			}
			w.WriteHeader(400)
			dat, _ := json.Marshal(responseError)
			w.Write(dat)
			return
		}
		if len(params.Body) >140 {
			responseError := returnVals{
				Error: "Chirp is too long",
			}
			w.WriteHeader(400)
			dat, _ := json.Marshal(responseError)
			w.Write(dat)
			return
		}
		response := returnVals{
			Valid: true,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		dat, _ := json.Marshal(response)
		w.Write(dat)
		
	})
	srv := &http.Server{
		Addr: ":" + port,
		Handler: mux,
	}
	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
//go build -o out && ./out
