package raft

type handlerFn func(n *Node,evt interface{})

type handlers struct {
	functions map[Role]handlerFn
}

func NewHandlers() *handlers {
	h := new(handlers)
	h.functions = make(map[Role]handlerFn)

	h.functions[Follower] = followerFn
	h.functions[Leader] = leaderFn
	h.functions[Candidate] = candidateFn

	return h
}



