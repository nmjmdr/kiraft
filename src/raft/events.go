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
