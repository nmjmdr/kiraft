package raft

import (
	"time"
)

type ElectionNotice struct {
	t time.Time
}

type TimeForHeartbeat struct {
	t time.Time
}

type StartElection struct {
}

type GotVote struct {
	response VoteResponse
}

type StartLeader struct {	
}


type GotAppendEntryResponse struct {
}
