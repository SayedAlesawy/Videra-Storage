package jobscheduler

// hasPostExecution checks if a job has a post job
func hasPostExecution(postExecution postJob) bool {
	return postExecution.ID != ""
}
