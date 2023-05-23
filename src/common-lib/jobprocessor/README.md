# Jobprocessor

-Package jobprocessor contains logic to create pool of goroutines to run any number of tasks parallely.This helps in submitting job by passing struct objects.

-Any consumer would need to implements Work() error function and need to submit a job
by passing the object of the struct which implements this method.

# Initialize a workerpool with any number of goroutines:
- It creates the workerpool using the passed jobContext.A JobContext mostly contains:
	- timeout-timeout is max time allowed to each worker for run
	- PoolSize- WorkRoutineCount is count of workers the consumer want in workerpool
	- TransactionID - TransactionID is the id for each job transaction to uniquely identify workers of that job.
- When the workerpool is initialized it creates channels to recive jobs submiited by consumers
and then invokes the Work method on the passed object.

# Usage:

package main

import (
	worker "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/jobprocessor"
)

type Products struct {
	name  string

}

func(p *Products) Work(id int) error{
	write code to execute custom logic
	return nil
}

func main(){
	pool,_:= worker.Initialize(JobContext{WorkRoutineCount:20,timeout:10,TransactionID:"trx123"})
	productList:=[]string{"product1","product2","product3","product4"}
	for _, pname := range productList{
		product:=Products{
			name:pname,
		}
		pool.AddJob(&product)
	}
	pool.Shutdown()


}
