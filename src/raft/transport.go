package raft

type Transport interface {
	RequestForVote(vreq VoteRequest,peer Peer) (VoteResponse,error)
	AppendEntry(entry Entry,peer Peer) (AppendResponse,error)
}


