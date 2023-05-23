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

// getCommitOffset returns the offset to commit
func (p *partitionState) getCommitOffset(offset int64) (int64, bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	size := len(p.offset)
	sort.Sort(p)

	var commitStatus *offsetStatus
	var commitIndex int
	var inProgressFound = false
	for index := 0; index < size; index++ {
		offsetStatus := p.offset[index]

		// update offset to complete
		if offsetStatus.offset == offset {
			p.updateOffset(index, false)
		}

		// if not completed and there is a greater "inProgress" offset, do nothing because that will commit when its done
		if offsetStatus.status != completed {
			inProgressFound = true
			if offsetStatus.offset > offset {
				break
			}
			continue
		}

		// set to the greatest offset available to be committed
		if !inProgressFound {
			commitStatus = offsetStatus
			commitIndex = index
		}
	}

	// update offsets in store and return the offset to commit
	if commitStatus != nil && commitStatus.offset > p.lastCommitedOffset {
		p.lastCommitedOffset = commitStatus.offset
		p.updateOffset(commitIndex, true)
		return commitStatus.offset, true
	}
	return -1, false
}

func (p *partitionState) updateOffset(index int, updateSlice bool) {
	p.offset[index].status = completed
	if updateSlice {
		if index < len(p.offset) {
			p.offset = p.offset[index+1:]
		}
	}
}

func (p *partitionState) setOffset(offset []*offsetStatus) {
	p.mutex.Lock()
	p.offset = offset
	p.mutex.Unlock() // Not using defer to avoid delay in unlocking
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
