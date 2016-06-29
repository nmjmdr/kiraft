package main


import (
	"encoding/json"
	"fmt"
	"net/http"
	//"os"
	"raft"
	"strings"
	"github.com/gorilla/mux"
	"strconv"
)


type State struct {	
	Node   string `json:"node-id"`
	Role   string `json:"role"`
	IsRunning bool `json:"is-running"`
}

type InitParams struct {
	NumberOfNodes int `json:"num-nodes"`
}

type RunFlag struct {
	IsRunning bool `json:"is-running"`
}

func (h *HttpServer) startSimulator(w http.ResponseWriter, r *http.Request) {

	// read number of nodes from body
	decoder := json.NewDecoder(r.Body)

	var p InitParams
	err := decoder.Decode(&p)

	fmt.Println("\nStart Simulator invoked")
	
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Could not serialize request body. Include num-nodes in the request body.  Error: %s", err)
		return
	}

	if p.NumberOfNodes <=0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "num-nodes should be greater than 0")
		return
	}

	h.s = raft.NewSimulator(p.NumberOfNodes)
	h.s.Start()
	
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Simulation created")
}


func (h *HttpServer) stopSimulator(w http.ResponseWriter, r *http.Request) {

	h.s.Stop()
}


func (h *HttpServer) getStates(w http.ResponseWriter, r *http.Request) {

	strStates := h.s.GetStates()

	fmt.Println(strStates)
	
	states := make([]State,len(strStates))

	for index,s := range strStates {
		splits := strings.Split(s,":")
		states[index].Node = splits[0]
		nodeIndex,_ := strconv.Atoi(splits[0])
		running,_ := h.s.IsRunning(nodeIndex)
		if running {
			states[index].Role = getRoleString(splits[1])
			states[index].IsRunning = true
		} else {
			states[index].Role = " - "
			states[index].IsRunning = false
		}
	}

	fmt.Println(states)

	responseBody,_ := json.MarshalIndent(states,"","    ")
	fmt.Fprintf(w,string(responseBody))
}

func getRoleString(p string) string {

	switch p {
	case "1":
		return "Leader"
	case "2":
		return "Candidate"
	case "3":
		return "Follower"
	default:
		panic("Recived a status which is not a follower,candidate or leader")
	}
}

func (h *HttpServer) toggleRunning(w http.ResponseWriter, r *http.Request) {

	var p RunFlag
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&p)

	fmt.Println("\ntoggle running invoked")
	
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Could not serialize request body. Include is-running:true|false in the request body.  Error: %s", err)
		return
	}

	vars := mux.Vars(r)
	nodeIndexStr,ok := vars["node-index"]

	nodeIndex,convertErr := strconv.Atoi(nodeIndexStr)

	if convertErr != nil || !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Could not get node-index from URI")
		return
	}	

	numNodes := h.s.NumNodes()
	if nodeIndex < 0 || nodeIndex > (numNodes -1) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "node-index should be between: 0 and %d",numNodes-1)
		return
	}

	currentlyRunning,_ := h.s.IsRunning(nodeIndex)

	if p.IsRunning == true {
		// start the node
		h.setAsStarted(nodeIndex,currentlyRunning,w)
	} else {
		h.setAsStopped(nodeIndex, currentlyRunning,w)
	}
}


func (h *HttpServer) setAsStarted(nodeIndex int,currentlyRunning bool,w http.ResponseWriter) {

	if currentlyRunning {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Node: %d is already running",nodeIndex)
		return
	}
	

	err := h.s.StartNode(nodeIndex)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Could not start node: %d, Error: %s",nodeIndex,err)
		return
	}
}


func (h *HttpServer) setAsStopped(nodeIndex int,currentlyRunning bool,w http.ResponseWriter) {

	if !currentlyRunning {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Node: %d is already stopped",nodeIndex)
		return
	}
	

	err := h.s.StopNode(nodeIndex)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Could not stop node: %d, Error: %s",nodeIndex,err)
		return
	}
}
