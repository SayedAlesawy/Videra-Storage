package jobscheduler

// JobQueue is responsible for scheduling jobs
type JobQueue struct {
	jobsQueue chan job      //jobs queue to be run
	tokens    chan struct{} //represents available slots for job
	capacity  int           //maximum number of concurrent jobs
}

type job struct {
	name string
	dir  string
	cmd  string
	args []string
}
