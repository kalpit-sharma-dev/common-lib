package consumer

import (
	"fmt"
	"sort"
	"sync"
)

const (
	inProgress = -1
	completed  = -2
)

var status2Name = map[int]string{inProgress: "In-Progress", completed: "Completed"}

type offsetStatus struct {
	offset int64
	status int
}

func (o *offsetStatus) String() string {
	return fmt.Sprintf("%s/%d", status2Name[o.status], o.offset)
}

type partitionState struct {
	offset             []*offsetStatus
	lastCommitedOffset int64
	dirty              bool
	mutex              sync.Mutex
}

// Len is part of sort.Interface.
func (p *partitionState) Len() int {
	return len(p.offset)
}

// Swap is part of sort.Interface.
func (p *partitionState) Swap(i, j int) {
	p.offset[i], p.offset[j] = p.offset[j], p.offset[i]
}

//Less is part of sort.Interface
func (p *partitionState) Less(i, j int) bool {
	return p.offset[i].offset < p.offset[j].offset
}

func (p *partitionState) getCommitOffset(offset int64) (int64, bool) {
	size := len(p.offset)
	p.mutex.Lock()
	sort.Sort(p)
	p.mutex.Unlock() //Not using defer to avoid delay in unlocking

	var commitStatus *offsetStatus
	var commitIndex int
	var inProgressFound = false
	for index := 0; index < size; index++ {
		offsetStatus := p.offset[index]

		if offsetStatus.offset == offset {
			p.updateOffset(index, false)
		}

		if offsetStatus.status != completed {
			inProgressFound = true
			if offsetStatus.offset > offset {
				break
			}
			continue
		}
		if !inProgressFound {
			commitStatus = offsetStatus
			commitIndex = index
		}
	}

	if commitStatus != nil && commitStatus.offset > p.lastCommitedOffset {
		p.lastCommitedOffset = commitStatus.offset
		p.updateOffset(commitIndex, true)
		return commitStatus.offset, true
	}
	return -1, false
}

func (p *partitionState) updateOffset(index int, updateSlice bool) {
	p.mutex.Lock()
	p.offset[index].status = completed
	if updateSlice {
		p.offset = p.offset[index+1:]
	}
	p.mutex.Unlock() //Not using defer to avoid delay in unlocking
}

func (p *partitionState) setOffset(offset []*offsetStatus) {
	p.mutex.Lock()
	p.offset = offset
	p.mutex.Unlock() //Not using defer to avoid delay in unlocking
}

func (p *partitionState) updateStatus(offset int64) {
	p.mutex.Lock()
	for _, value := range p.offset {
		if value.offset == offset {
			value.status = completed
			p.dirty = true
			break
		}
	}
	p.mutex.Unlock() //Not using defer to avoid delay in unlocking
}
