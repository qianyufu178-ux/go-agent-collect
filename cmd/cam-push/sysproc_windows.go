//go:build windows

package main

import "syscall"

// hiddenProcAttr 让 ffmpeg 子进程不弹出黑色控制台窗口。
func hiddenProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
}
