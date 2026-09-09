//go:build windows

package app

import "syscall"

const detachedTrayFlags = syscall.CREATE_NEW_PROCESS_GROUP | 0x00000008
