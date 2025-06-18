package kvstore

type KVStore interface {
	// Methods for accessing cached PagerDuty data
	GetCachedSchedules() ([]byte, error)
	SetCachedSchedules(data []byte) error
}
