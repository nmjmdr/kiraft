package main

import (
	//"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	//"os"
	"raft"
)



type HttpServer struct {
	routeMap map[string]map[string]http.HandlerFunc
	s raft.Simulator
}


func (h *HttpServer) HandleRequests(port string) {

	router := mux.NewRouter().StrictSlash(true)
	// calling .Methods("GET") returns 404 not found, where as it should be returning 405 method not allowed
	// hence the wrapper function to return 405 in code
	router.HandleFunc("/simulator", h.filter("/simulator"))
	router.HandleFunc("/simulator/states", h.filter("/simulator/states"))
	router.HandleFunc("/simulator/nodes/{node-index}/running", h.filter("/simulator/nodes/{node-index}/running"))

	port = ":" + port
	log.Fatal(http.ListenAndServe(port, router))
}



func (h *HttpServer) describe(w http.ResponseWriter, r *http.Request) {	
	fmt.Fprintf(w, "description")
}

func (h *HttpServer) filter(path string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}		
		fn := h.dispatch(r.Method,path)
		if fn != nil {
			fn(w,r)
		} else {
			// method not supported			
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, "%s not allowed on path %s", r.Method,path)
		}
	}
}


func (h *HttpServer) dispatch(method string,path string) http.HandlerFunc {

	switch method {
	case "OPTIONS":
		return h.describe
	default:
		fmt.Printf("Method: %s, path: %s\n",method,path)
		fmt.Println(h.routeMap[method][path])
		return h.routeMap[method][path]
	}
}

func (h *HttpServer) setupRoutes() {
	h.routeMap["GET"] = make(map[string]http.HandlerFunc)
	h.routeMap["PUT"] = make(map[string]http.HandlerFunc)
	h.routeMap["POST"] = make(map[string]http.HandlerFunc)
	h.routeMap["DELETE"] = make(map[string]http.HandlerFunc)

	h.routeMap["POST"]["/simulator"] = h.startSimulator
	h.routeMap["DELETE"]["/simulator"] = h.stopSimulator
	h.routeMap["GET"]["/simulator/states"] = h.getStates
	h.routeMap["PUT"]["/simulator/nodes/{node-index}/running"] = h.toggleRunning
	
	
}

func NewHttpServer() *HttpServer {
	h := new(HttpServer)
	h.routeMap = make(map[string]map[string]http.HandlerFunc)
	h.setupRoutes()
	return h
}

