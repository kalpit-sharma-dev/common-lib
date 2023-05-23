package main

import (
	"time"

	testutils "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/testUtils"
)

type RepositoryUpdates struct {
	updatedBy string
	updatedAt time.Time
}

type CodeRepository struct {
	name string
	RepositoryUpdates
}

func main() {
	repositoryA := CodeRepository{
		name:              "repositoryA",
		RepositoryUpdates: RepositoryUpdates{"UserB", time.Now()},
	}
	repositoryB := CodeRepository{
		name:              "repositoryA",
		RepositoryUpdates: RepositoryUpdates{"UserA", time.Now().AddDate(0, 0, 1)},
	}
	_ = testutils.DeepEqualIgnoringTime(repositoryA, repositoryB)
}
