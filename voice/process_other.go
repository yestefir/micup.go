//go:build !windows

package voice

import "os/exec"

func configure(*exec.Cmd) {}
