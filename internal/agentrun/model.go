package agentrun

import "time"

type AgentRun struct {
	ID        string    `json:"id" xorm:"pk varchar(64)"`
	AgentID   string    `json:"agent_id" xorm:"varchar(128) index"`
	CaseID    string    `json:"case_id" xorm:"varchar(64) index"`
	Target    string    `json:"target" xorm:"text"`
	Status    string    `json:"status" xorm:"varchar(32) index"`
	CreatedAt time.Time `json:"created_at" xorm:"created"`
	UpdatedAt time.Time `json:"updated_at" xorm:"updated"`
}

func NewAgentRun(agentID, caseID, target string) *AgentRun {
	return &AgentRun{
		ID:       generateID(),
		AgentID:  agentID,
		CaseID:   caseID,
		Target:   target,
		Status:   "running",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func generateID() string {
	return "agent-run-" + time.Now().Format("20060102150405")
}
