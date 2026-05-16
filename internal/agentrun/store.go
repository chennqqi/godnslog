package agentrun

import "xorm.io/xorm"

type Store struct {
	engine *xorm.Engine
}

func NewStore(engine *xorm.Engine) *Store {
	return &Store{engine: engine}
}

func (s *Store) Create(run *AgentRun) error {
	_, err := s.engine.Insert(run)
	return err
}

func (s *Store) Get(id string) (*AgentRun, error) {
	run := new(AgentRun)
	found, err := s.engine.ID(id).Get(run)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return run, nil
}
