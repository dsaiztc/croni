package cmd

import (
	"time"

	"github.com/dsaiztc/croni/internal/launchd"
	"github.com/dsaiztc/croni/internal/store"
	"github.com/dsaiztc/croni/internal/types"
)

func reapExpiredAtJobs(s *store.Store) {
	jobs, err := s.List()
	if err != nil {
		return
	}

	var expired []string
	for _, j := range jobs {
		if j.Schedule.Type != types.ScheduleAt || !j.Enabled {
			continue
		}
		t, err := launchd.ParseAt(j.Schedule.Expression)
		if err != nil {
			continue
		}
		if t.Before(time.Now()) {
			launchd.Bootout(j.Name)
			launchd.RemovePlist(j.Name)
			expired = append(expired, j.Name)
		}
	}

	s.RemoveMany(expired)
}
