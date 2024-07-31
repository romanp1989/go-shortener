package storage

type CacheStorage struct {
	storageURL map[string]string
}

func NewCacheStorage() *CacheStorage {
	return &CacheStorage{storageURL: make(map[string]string)}
}

func (c *CacheStorage) Get(inputURL string) string {
	if foundurl, ok := c.storageURL[inputURL]; ok {
		return foundurl
	}
	return ""
}

func (c *CacheStorage) Save(originalURL string, shortURL string) error {
	c.storageURL[shortURL] = originalURL
	c.storageURL[originalURL] = shortURL
	return nil
}
