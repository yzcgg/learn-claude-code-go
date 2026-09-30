package permission

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

// 教学示意：字符串匹配不是可靠的安全机制，命令变体和 shell 展开可能绕过。
var denyList = []string{
	"rm -rf /", "sudo", "shutdown", "reboot",
	"mkfs", "dd if=", "> /dev/sda",
}

type permissionRule struct {
	tools   []string
	check   func(args map[string]any, workDir string) bool
	message string
}

var permissionRules = []permissionRule{
	{
		tools: []string{"read_file", "write_file", "edit_file", "create_file"},
		check: func(args map[string]any, workDir string) bool {
			path, _ := args["path"].(string)
			return outsideWorkspace(path, workDir)
		},
		message: "Access outside workspace",
	},
	{
		tools: []string{"bash"},
		check: func(args map[string]any, _ string) bool {
			command, _ := args["command"].(string)
			for _, keyword := range []string{"rm ", "> /etc/", "chmod 777"} {
				if strings.Contains(command, keyword) {
					return true
				}
			}
			return false
		},
		message: "Potentially destructive command",
	},
}

// Reuse the reader so buffered input is preserved between approval prompts.
var approvalInput = bufio.NewReader(os.Stdin)

func CheckPermission(block anthropic.ToolUseBlock, workDir string) bool {
	var args map[string]any
	if err := json.Unmarshal(block.Input, &args); err != nil || args == nil {
		fmt.Println("\n⛔ Blocked: invalid tool input")
		return false
	}

	// 闸门 1：硬拒绝优先，不能通过用户审批覆盖。
	if block.Name == "bash" {
		command, _ := args["command"].(string)
		if reason := checkDenyList(command); reason != "" {
			fmt.Printf("\n⛔ %s\n", reason)
			return false
		}
	}

	// 闸门 2 + 3：规则命中后，等待用户审批。
	if reason := checkRules(block.Name, args, workDir); reason != "" {
		return askUser(block.Name, args, reason)
	}

	return true
}

func checkDenyList(command string) string {
	for _, pattern := range denyList {
		if strings.Contains(command, pattern) {
			return fmt.Sprintf("Blocked: '%s' is on the deny list", pattern)
		}
	}
	return ""
}

func checkRules(toolName string, args map[string]any, workDir string) string {
	for _, rule := range permissionRules {
		if slices.Contains(rule.tools, toolName) && rule.check(args, workDir) {
			return rule.message
		}
	}
	return ""
}

func askUser(toolName string, args map[string]any, reason string) bool {
	input, err := json.Marshal(args)
	if err != nil {
		return false
	}
	fmt.Printf("\n⚠  %s\n   Tool: %s(%s)\n   Allow? [y/N] ", reason, toolName, input)
	choice, err := approvalInput.ReadString('\n')
	if err != nil {
		return false
	}
	choice = strings.ToLower(strings.TrimSpace(choice))
	return choice == "y" || choice == "yes"
}

func outsideWorkspace(path string, workDir string) bool {
	if !filepath.IsAbs(path) {
		path = filepath.Join(workDir, path)
	}
	root, err := resolvePath(workDir)
	if err != nil {
		return true
	}
	target, err := resolvePath(path)
	if err != nil {
		return true
	}
	rel, err := filepath.Rel(root, target)
	return err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// Resolve existing symlinks while allowing a missing file or parent directory.
// Resolution errors trigger approval instead of silently allowing access.
func resolvePath(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err == nil {
		return resolved, nil
	}
	if !os.IsNotExist(err) {
		return "", err
	}
	// An existing dangling symlink cannot be treated as a missing regular file.
	if _, statErr := os.Lstat(absolute); statErr == nil || !os.IsNotExist(statErr) {
		return "", err
	}
	parent := filepath.Dir(absolute)
	if parent == absolute {
		return "", err
	}
	resolvedParent, err := resolvePath(parent)
	if err != nil {
		return "", err
	}
	return filepath.Join(resolvedParent, filepath.Base(absolute)), nil
}
