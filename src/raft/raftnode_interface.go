package raft

type RaftNode interface {
	Start()
	Stop()
	CurrentRole() Role
	IsRunning() bool
	AppendEntry(Entry) (AppendResponse,error)
	RequestForVote(VoteRequest) (VoteResponse,error)
	RoleChange() (chan Role)
	Id() string
}


type VoteRequest struct {
	Term uint64
	CandidateId string
	LastLogIndex uint64
	LastLogTerm uint64
}

type VoteResponse struct {
	TermToUpdate uint64
	VoteGranted bool
	From string
}

type Entry struct {
	Term uint64
	From string
}

type AppendResponse struct {
	Reply bool
	Term uint64
}

