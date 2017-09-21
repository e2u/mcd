// 用文件形式存储远程资源
package cache

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/e2u/goboot"
)

type FileStorage struct {
	sync.RWMutex
	path string
}

func NewFileStorage(path string) *FileStorage {
	return &FileStorage{
		path: path,
	}
}

// 对 key 做格式化,传入的 key 是原始的 url,需要转换成合适的路径
// 存储路径格式转换算法,用 key 计算出 md5
func (fs *FileStorage) fullPath(key string) string {
	hash := fmt.Sprintf("%x", md5.Sum([]byte(key)))
	return func() string {
		p, _ := filepath.Abs(fmt.Sprintf("%s/%s/%s/%s", fs.path, hash[31:], hash[29:31], hash))
		return p
	}()
}

func (fs *FileStorage) mkdirs(fullpath string) error {
	dirpath := filepath.Dir(fullpath)
	if _, err := os.Stat(dirpath); err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(dirpath, 0755)
		}
	} else {
		return err
	}
	return nil
}

func (fs *FileStorage) Get(key string) (*CacheObject, error) {
	fs.RLock()
	defer fs.RUnlock()

	f, err := os.Open(fs.fullPath(key))
	if err != nil {
		return nil, err
	}

	defer f.Close()
	return LoadCacheObject(f)
}

func (fs *FileStorage) Set(key string, oc *CacheObject) error {
	goboot.Log.Debugf("Set %s", key)
	fs.Lock()
	defer fs.Unlock()
	fullpath := fs.fullPath(key)

	if err := fs.mkdirs(fullpath); err != nil {
		return err
	}

	f, err := os.Create(fullpath)
	if err != nil {
		return err
	}
	defer f.Close()
	return DumpCacheObject(f, oc)
}

func (fs *FileStorage) Delete(key string) error {
	goboot.Log.Debugf("Delete %s", key)
	fs.Lock()
	defer fs.Unlock()
	fullpath := fs.fullPath(key)
	return os.Remove(fullpath)
}
