package jobscheduler

import (
	"sync"

	"github.com/SayedAlesawy/Videra-Storage/config"
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
		}

		jobQueueInstance = &jobQueue
	})

	return jobQueueInstance
}
