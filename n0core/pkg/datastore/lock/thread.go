package lock

type mutexAction int

const (
	lock     mutexAction = iota
	unlock   mutexAction = iota
	isLocked mutexAction = iota
)

type mutexRequest struct {
	Key    string
	Action mutexAction
	Result chan mutexResult
}

type mutexResult struct {
	Succeeded bool
	Locked    bool
}

func (mt *MemoryMutexTable) mutexThread() {
	for req := range mt.request {
		succeeded := false

		switch req.Action {
		case lock:
			succeeded = mt.lock(req.Key)

		case unlock:
			succeeded = mt.unlock(req.Key)

		case isLocked:
			succeeded = true
		}

		req.Result <- mutexResult{
			Succeeded: succeeded,
			Locked:    mt.isLocked(req.Key),
		}
	}
}

func (mt *MemoryMutexTable) lock(key string) bool {
	if mt.isLocked(key) {
		return false
	}

	// raft consensus
	mt.table[key] = true

	return true
}

func (mt *MemoryMutexTable) unlock(key string) bool {
	if !mt.isLocked(key) {
		return false
	}

	// raft consensus
	delete(mt.table, key)

	return true
}

func (mt MemoryMutexTable) isLocked(key string) bool {
	if _, ok := mt.table[key]; ok {
		return true
	}

	return false
}
