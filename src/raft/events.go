package raft

import (
	//"time"
)

type ElectionNotice struct {
}

type TimeForHeartbeat struct {
}

type StartElection struct {
}

type GotVote struct {
	response VoteResponse
}

type StartLeader struct {	
}


type GotAppendEntryResponse struct {
	response AppendResponse
}


type HigherTermDiscovered struct {
	term uint64
	from string
}

type GotAppendEntryRequest struct {
	entry Entry
}

type GotRequestForVote struct {
	voteRequest VoteRequest
}

