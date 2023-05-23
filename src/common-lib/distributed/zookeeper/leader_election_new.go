package zookeeper

import (
	"fmt"
	"strings"
	"time"

	election "github.com/Comcast/go-leaderelection"
	"github.com/samuel/go-zookeeper/zk"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/distributed/leader-election"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/sync"
)

const pathAlreadyExist = "node already exists"

type (
	// service struct
	electionImpl struct {
		candidate *election.Election
		zkConn    *zk.Conn
		config    *sync.Config
	}

	electionResponse struct {
		IsLeader    bool
		CandidateID string
	}
)

// GetService is a function to return service instance
func GetService(config *sync.Config) leaderelection.Interface {
	return &electionImpl{config: config}
}

// RegisterCandidate : Register the Candidate for election
func (e *electionImpl) RegisterCandidate(electionResource string, clientName string) (err error) {
	if e.zkConn == nil {
		e.zkConn, _, err = zk.Connect(e.config.Servers, time.Duration(e.config.SessionTimeoutInSecond)*time.Second)
		if err != nil {
			return err
		}
	}

	_, err = e.zkConn.Create(electionResource, []byte(""), 0, zk.WorldACL(zk.PermAll))
	if err != nil && !strings.Contains(err.Error(), pathAlreadyExist) {
		return err
	}

	e.candidate, err = election.NewElection(e.zkConn, electionResource, clientName)
	if err != nil {
		return err
	}
	return nil
}

// ResignCandidate : Resign Candidate from election
func (e *electionImpl) ResignCandidate() {
	e.candidate.Resign()
	e.zkConn.Close()
}

// StartElection : Start the election and call the callback function if node win the election
func (e *electionImpl) StartElection(callback func()) error {

	go e.candidate.ElectLeader()

	var status election.Status
	var ok bool
	respCh := make(chan electionResponse)
	connFailCh := make(chan bool)

	for {
		select {
		case status, ok = <-e.candidate.Status():
			if !ok {
				// Channel closed, election is terminated !!
				respCh <- electionResponse{false, status.CandidateID}
				e.ResignCandidate()
				return fmt.Errorf("ErrChannelClosed")
			}
			if status.Err != nil {
				// Got error in election
				e.ResignCandidate()
				return fmt.Errorf("ErrReceivedElectionStatusError : Candidate : %s : Error : %v ", status.CandidateID, status.Err)
			}
			if status.Role == election.Leader {
				callback()
				e.ResignCandidate()
				return nil
			}
		case <-connFailCh:
			respCh <- electionResponse{false, status.CandidateID}
			e.ResignCandidate()
			return fmt.Errorf("ErrZookeeperConnectionFailed : candidate : %s", status.CandidateID)
		}
	}
}

// BecomeALeader : Old implementation : ignore for this file
func (e *electionImpl) BecomeALeader() (int, bool, error) {
	return 0, false, nil
}
