package cache

type Storager interface {
	Get(key string) (*CacheObject, error)
	Set(key string, oc *CacheObject) error
	Delete(key string) error
}
