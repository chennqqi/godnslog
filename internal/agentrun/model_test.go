package agentrun

import "testing"

func TestAgentRunDefaultsToRunning(t *testing.T) {
	run := NewAgentRun("agent-1", "case-1", "https://target.example")
	if run.Status != "running" {
		t.Fatalf("expected running status, got %s", run.Status)
	}
	if run.AgentID != "agent-1" || run.CaseID != "case-1" {
		t.Fatalf("unexpected agent run identity: %#v", run)
	}
}
