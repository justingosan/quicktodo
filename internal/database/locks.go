package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// LockManager manages file locks for database operations
type LockManager struct {
	lockDir string
	timeout time.Duration
}

// NewLockManager creates a new lock manager
func NewLockManager(lockDir string, timeoutSeconds int) *LockManager {
	return &LockManager{
		lockDir: lockDir,
		timeout: time.Duration(timeoutSeconds) * time.Second,
	}
}

// LockInfo contains information about a lock
type LockInfo struct {
	ProcessID int
	CreatedAt time.Time
	FilePath  string
}

// AcquireLock attempts to acquire a lock for the given project
func (lm *LockManager) AcquireLock(projectName string) (*LockInfo, error) {
	// Ensure lock directory exists
	if err := os.MkdirAll(lm.lockDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create lock directory: %w", err)
	}

	lockPath := filepath.Join(lm.lockDir, projectName+".lock")

	// Check for existing lock
	if existingLock, err := lm.readLockFile(lockPath); err == nil {
		// Check if the lock is stale
		if time.Since(existingLock.CreatedAt) > 5*time.Minute {
			// Remove stale lock
			if err := os.Remove(lockPath); err != nil {
				return nil, fmt.Errorf("failed to remove stale lock: %w", err)
			}
		} else {
			// Check if process is still running
			if lm.isProcessRunning(existingLock.ProcessID) {
				return nil, fmt.Errorf("project %s is locked by process %d", projectName, existingLock.ProcessID)
			} else {
				// Process is dead, remove lock
				if err := os.Remove(lockPath); err != nil {
					return nil, fmt.Errorf("failed to remove orphaned lock: %w", err)
				}
			}
		}
	}

	// Create new lock
	lockInfo := &LockInfo{
		ProcessID: os.Getpid(),
		CreatedAt: time.Now(),
		FilePath:  lockPath,
	}

	// Try to acquire lock with timeout
	startTime := time.Now()
	for time.Since(startTime) < lm.timeout {
		if err := lm.writeLockFile(lockPath, lockInfo); err == nil {
			return lockInfo, nil
		}

		// Wait a bit before retrying
		time.Sleep(100 * time.Millisecond)
	}

	return nil, fmt.Errorf("timeout acquiring lock for project %s", projectName)
}

// ReleaseLock releases a lock
func (lm *LockManager) ReleaseLock(lockInfo *LockInfo) error {
	if lockInfo == nil || lockInfo.FilePath == "" {
		return fmt.Errorf("invalid lock info")
	}

	// Verify we own the lock
	if currentLock, err := lm.readLockFile(lockInfo.FilePath); err != nil {
		// Lock file doesn't exist, consider it released
		return nil
	} else if currentLock.ProcessID != lockInfo.ProcessID {
		return fmt.Errorf("lock is owned by different process")
	}

	// Remove lock file
	if err := os.Remove(lockInfo.FilePath); err != nil {
		return fmt.Errorf("failed to remove lock file: %w", err)
	}

	return nil
}

// readLockFile reads lock information from file
func (lm *LockManager) readLockFile(lockPath string) (*LockInfo, error) {
	data, err := os.ReadFile(lockPath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("invalid lock file format")
	}

	processID, err := strconv.Atoi(strings.TrimSpace(lines[0]))
	if err != nil {
		return nil, fmt.Errorf("invalid process ID in lock file")
	}

	createdAt, err := time.Parse(time.RFC3339, strings.TrimSpace(lines[1]))
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp in lock file")
	}

	return &LockInfo{
		ProcessID: processID,
		CreatedAt: createdAt,
		FilePath:  lockPath,
	}, nil
}

// writeLockFile writes lock information to file
func (lm *LockManager) writeLockFile(lockPath string, lockInfo *LockInfo) error {
	// Use O_EXCL to ensure atomic creation
	file, err := os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	content := fmt.Sprintf("%d\n%s\n", lockInfo.ProcessID, lockInfo.CreatedAt.Format(time.RFC3339))
	_, err = file.WriteString(content)
	return err
}

// isProcessRunning checks if a process is still running
func (lm *LockManager) isProcessRunning(pid int) bool {
	// Send signal 0 to check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// CleanupStaleLocks removes stale locks older than the specified duration
func (lm *LockManager) CleanupStaleLocks(maxAge time.Duration) ([]string, error) {
	var cleaned []string

	// List all lock files
	files, err := os.ReadDir(lm.lockDir)
	if err != nil {
		if os.IsNotExist(err) {
			return cleaned, nil
		}
		return nil, fmt.Errorf("failed to read lock directory: %w", err)
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".lock") {
			continue
		}

		lockPath := filepath.Join(lm.lockDir, file.Name())
		lockInfo, err := lm.readLockFile(lockPath)
		if err != nil {
			// Remove invalid lock file
			if err := os.Remove(lockPath); err == nil {
				cleaned = append(cleaned, file.Name())
			}
			continue
		}

		// Check if lock is stale
		if time.Since(lockInfo.CreatedAt) > maxAge {
			if err := os.Remove(lockPath); err == nil {
				cleaned = append(cleaned, file.Name())
			}
		} else if !lm.isProcessRunning(lockInfo.ProcessID) {
			// Process is dead, remove lock
			if err := os.Remove(lockPath); err == nil {
				cleaned = append(cleaned, file.Name())
			}
		}
	}

	return cleaned, nil
}

// GetActiveLocks returns information about all active locks
func (lm *LockManager) GetActiveLocks() (map[string]*LockInfo, error) {
	locks := make(map[string]*LockInfo)

	// List all lock files
	files, err := os.ReadDir(lm.lockDir)
	if err != nil {
		if os.IsNotExist(err) {
			return locks, nil
		}
		return nil, fmt.Errorf("failed to read lock directory: %w", err)
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".lock") {
			continue
		}

		lockPath := filepath.Join(lm.lockDir, file.Name())
		lockInfo, err := lm.readLockFile(lockPath)
		if err != nil {
			continue
		}

		// Only include active locks (process still running)
		if lm.isProcessRunning(lockInfo.ProcessID) {
			projectName := strings.TrimSuffix(file.Name(), ".lock")
			locks[projectName] = lockInfo
		}
	}

	return locks, nil
}

// ForceLock forcefully acquires a lock by removing any existing lock
func (lm *LockManager) ForceLock(projectName string) (*LockInfo, error) {
	lockPath := filepath.Join(lm.lockDir, projectName+".lock")

	// Remove existing lock if present
	if err := os.Remove(lockPath); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to remove existing lock: %w", err)
	}

	// Create new lock
	lockInfo := &LockInfo{
		ProcessID: os.Getpid(),
		CreatedAt: time.Now(),
		FilePath:  lockPath,
	}

	if err := lm.writeLockFile(lockPath, lockInfo); err != nil {
		return nil, fmt.Errorf("failed to create lock: %w", err)
	}

	return lockInfo, nil
}
