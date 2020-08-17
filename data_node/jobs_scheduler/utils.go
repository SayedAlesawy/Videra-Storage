package jobscheduler

// hasPostExecution checks if a job has a post job
func hasPostExecution(postExecution PostJob) bool {
	return postExecution.TableName != ""
}
