package jobscheduler

import "time"

// JobQueue is responsible for scheduling jobs
type JobQueue struct {
	jobsQueue chan job      //jobs queue to be run
	tokens    chan struct{} //represents available slots for job
	capacity  int           //maximum number of concurrent jobs
	timeout   time.Duration //time out for executing job
}

type job struct {
	name          string
	dir           string
	cmd           string
	args          []string
	postExecution postJob //update set after job execution
}

// postJob represents an update set to db after job execution
type postJob struct {
	ID         string
	TableName  string
	ColumnName string
	NewValue   string
}
