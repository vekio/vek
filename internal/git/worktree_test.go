package git

import "testing"

func TestParseWorktrees(t *testing.T) {
	output := "worktree /repo/.git\x00bare\x00\x00" +
		"worktree /repo/main\x00HEAD abc123\x00branch refs/heads/main\x00\x00" +
		"worktree /repo/feature\n42\x00HEAD def456\x00detached\x00\x00" +
		"worktree /repo/old\x00HEAD abc123\x00branch refs/heads/old\x00prunable gitdir file points to non-existent location\x00\x00"
	worktrees, err := ParseWorktrees(output)
	if err != nil {
		t.Fatal(err)
	}
	if len(worktrees) != 4 {
		t.Fatalf("worktrees = %+v", worktrees)
	}
	if !worktrees[0].Bare || worktrees[0].Path != "/repo/.git" {
		t.Fatalf("bare entry = %+v", worktrees[0])
	}
	if worktrees[1].Branch != "main" || worktrees[1].Bare {
		t.Fatalf("main entry = %+v", worktrees[1])
	}
	if !worktrees[2].Detached || worktrees[2].Path != "/repo/feature\n42" {
		t.Fatalf("detached entry = %+v", worktrees[2])
	}
	if !worktrees[3].Prunable || worktrees[3].Branch != "old" {
		t.Fatalf("prunable entry = %+v", worktrees[3])
	}
}

func TestParseWorktreesRejectsIncompleteOutput(t *testing.T) {
	for _, output := range []string{"worktree /repo/main\x00HEAD abc", "HEAD abc\x00\x00"} {
		if _, err := ParseWorktrees(output); err == nil {
			t.Errorf("ParseWorktrees(%q) succeeded", output)
		}
	}
}
