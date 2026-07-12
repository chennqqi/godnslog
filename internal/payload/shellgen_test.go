package payload

import (
	"strings"
	"testing"
)

func TestGenerateShell(t *testing.T) {
	cmds := GenerateShell("10.0.0.1", "4444")
	if len(cmds) != 8 {
		t.Errorf("expected 8 commands, got %d", len(cmds))
	}

	for _, cmd := range cmds {
		if !strings.Contains(cmd.Command, "10.0.0.1") {
			t.Errorf("%s: missing IP", cmd.Name)
		}
		if !strings.Contains(cmd.Command, "4444") {
			t.Errorf("%s: missing port", cmd.Name)
		}
	}
}

func TestGenerateShellDifferentPort(t *testing.T) {
	cmds := GenerateShell("192.168.1.100", "1337")
	if len(cmds) != 8 {
		t.Errorf("expected 8 commands, got %d", len(cmds))
	}

	for _, cmd := range cmds {
		if !strings.Contains(cmd.Command, "1337") {
			t.Errorf("%s: missing port 1337 in %s", cmd.Name, cmd.Command[:min(50, len(cmd.Command))])
		}
	}
}

func TestGenerateShellIPv6(t *testing.T) {
	cmds := GenerateShell("::1", "8080")
	if len(cmds) != 8 {
		t.Errorf("expected 8 commands, got %d", len(cmds))
	}
}
