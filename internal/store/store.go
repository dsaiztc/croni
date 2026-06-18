package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/dsaiztc/croni/internal/types"
)

type Store struct {
	path string
}

func New() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home dir: %w", err)
	}
	dir := filepath.Join(home, ".croni")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create croni dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "logs"), 0755); err != nil {
		return nil, fmt.Errorf("create logs dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "generated"), 0755); err != nil {
		return nil, fmt.Errorf("create generated dir: %w", err)
	}
	return &Store{path: filepath.Join(dir, "jobs.json")}, nil
}

func (s *Store) Dir() string {
	return filepath.Dir(s.path)
}

func (s *Store) load() (*types.Store, error) {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return &types.Store{Version: 1, Jobs: make(map[string]types.Job)}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read jobs.json: %w", err)
	}
	var st types.Store
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, fmt.Errorf("parse jobs.json: %w", err)
	}
	if st.Jobs == nil {
		st.Jobs = make(map[string]types.Job)
	}
	return &st, nil
}

func (s *Store) save(st *types.Store) error {
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal jobs.json: %w", err)
	}
	return os.WriteFile(s.path, data, 0644)
}

func (s *Store) withLock(fn func(*types.Store) error) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}
	lockPath := s.path + ".lock"
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("open lock file: %w", err)
	}
	defer f.Close()
	defer os.Remove(lockPath)

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("acquire lock: %w", err)
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)

	st, err := s.load()
	if err != nil {
		return err
	}
	if err := fn(st); err != nil {
		return err
	}
	return s.save(st)
}

func (s *Store) Add(job types.Job) error {
	return s.withLock(func(st *types.Store) error {
		if _, exists := st.Jobs[job.Name]; exists {
			return fmt.Errorf("job %q already exists", job.Name)
		}
		st.Jobs[job.Name] = job
		return nil
	})
}

func (s *Store) Get(name string) (types.Job, error) {
	st, err := s.load()
	if err != nil {
		return types.Job{}, err
	}
	job, ok := st.Jobs[name]
	if !ok {
		return types.Job{}, fmt.Errorf("job %q not found", name)
	}
	return job, nil
}

func (s *Store) List() ([]types.Job, error) {
	st, err := s.load()
	if err != nil {
		return nil, err
	}
	jobs := make([]types.Job, 0, len(st.Jobs))
	for _, j := range st.Jobs {
		jobs = append(jobs, j)
	}
	return jobs, nil
}

func (s *Store) Update(job types.Job) error {
	return s.withLock(func(st *types.Store) error {
		if _, exists := st.Jobs[job.Name]; !exists {
			return fmt.Errorf("job %q not found", job.Name)
		}
		st.Jobs[job.Name] = job
		return nil
	})
}

func (s *Store) Remove(name string) error {
	return s.withLock(func(st *types.Store) error {
		if _, exists := st.Jobs[name]; !exists {
			return fmt.Errorf("job %q not found", name)
		}
		delete(st.Jobs, name)
		return nil
	})
}

func (s *Store) RemoveMany(names []string) error {
	if len(names) == 0 {
		return nil
	}
	return s.withLock(func(st *types.Store) error {
		for _, name := range names {
			delete(st.Jobs, name)
		}
		return nil
	})
}
