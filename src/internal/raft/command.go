package raft

type operation string

const (
	opPut    = "put"
	opDelete = "delete"
)

type command struct {
	Op    operation `json:"op"`
	Key   string    `json:"key"`
	Value string    `json:"value"`
}
