package jobscheduler

import (
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"

	"github.com/SayedAlesawy/Videra-Storage/config"
	datanode "github.com/SayedAlesawy/Videra-Storage/data_node"
)

// logPrefix Used for hierarchical logging
var logPrefix = "[Job-Queue]"

// jobqueueOnce Used to garauntee thread safety for singleton instances
var jobqueueOnce sync.Once

// jobQueueInstance A singleton instance of the jobQueue object
var jobQueueInstance *JobQueue

// JobQueueInstance A function to return a singleton jobQueue instance
func JobQueueInstance() *JobQueue {
	dataNodeConfig := config.ConfigurationManagerInstance("").DataNodeConfig()

	jobqueueOnce.Do(func() {
		capacity := dataNodeConfig.MaximumConcurrentJobs
		jobQueue := JobQueue{
			jobsQueue: make(chan job),
			tokens:    make(chan struct{}, capacity),
			capacity:  capacity,
			timeout:   time.Duration(dataNodeConfig.JobTimeout) * time.Second,
		}
		jobQueue.addTokens(capacity)
		jobQueueInstance = &jobQueue
		go jobQueue.processJobs()
	})

	return jobQueueInstance
}

// InsertJob inserts a job into job queue to be executed
func (jobQueue *JobQueue) InsertJob(name string, cmd string, args []string, postExecution postJob) {
	jobQueue.InsertJobWithDir(name, "", cmd, args, postExecution)
}

// InsertJobWithDir inserts a job into job queue to be executed at dir
func (jobQueue *JobQueue) InsertJobWithDir(name string, dir string, cmd string, args []string, postExecution postJob) {
	jobQueue.jobsQueue <- job{name: name, dir: dir, cmd: cmd, args: args, postExecution: postExecution}
}

// processJobs is responsible for periodically process jobs from job queue
func (jobQueue *JobQueue) processJobs() {
	// Blocks untill there's a token
	// then if there's a token, block untill there's a job to be executed
	for {
		<-jobQueue.tokens
		nextJob := <-jobQueue.jobsQueue
		jobQueue.executeJob(nextJob)
	}
}

// executeJob is responsible for executing jobs inside the job queue
func (jobQueue *JobQueue) executeJob(executedJob job) {
	// add a token when job is finished
	defer jobQueue.addTokens(1)

	log.Println(logPrefix, "Starting executing job", executedJob.name)

	cmd := exec.Command(executedJob.cmd, executedJob.args...)
	if executedJob.dir != "" {
		cmd.Dir = executedJob.dir
	}

	err := cmd.Start()
	if err != nil {
		log.Println(logPrefix, "Error starting job", executedJob.name, err)
		return
	}

	done := make(chan error)
	timer := time.NewTimer(jobQueue.timeout)
	defer timer.Stop()

	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-done:
		log.Println(logPrefix, fmt.Sprintf("Job %s completed successfuly!", executedJob.name))
		if hasPostExecution(executedJob.postExecution) {
			err = updateDB(executedJob.postExecution)
			if err != nil {
				log.Println(logPrefix, fmt.Sprintf("Update DB for job %s failed", executedJob.name))
			}
		}
	case <-timer.C:
		cmd.Process.Kill()
		log.Println(logPrefix, fmt.Sprintf("Job %s timedout!", executedJob.name))
	}

}

// addTokens is responsible for adding tokens to jobQueue
func (jobQueue *JobQueue) addTokens(count int) {
	for count > 0 {
		jobQueue.tokens <- struct{}{}
		count--
	}
}

// updateDB is responsible for updating DB
func updateDB(postExecution postJob) error {
	dn := datanode.NodeInstance()
	return dn.DB.Connection.Raw("UPDATE ? SET ? = ? WHERE id = ?", postExecution.TableName, postExecution.ColumnName, postExecution.NewValue, postExecution.ID).Error
}
