//go:build !windows

package main

import "syscall"

// hiddenProcAttr 非 Windows 平台无需隐藏窗口。
func hiddenProcAttr() *syscall.SysProcAttr { return nil }
