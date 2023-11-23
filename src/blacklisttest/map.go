package main  
  
import (  
 "fmt"  
 "sync"  
 "time"  
)  
  
type Cache struct {  
 maxBytes  int64  
 keys      map[string]interface{}  
 mutex     sync.RWMutex  
 evictTick *time.Ticker  
}  
  
func NewCache(maxBytes int64) *Cache {  
 return &Cache{  
 maxBytes:  maxBytes,  
 keys:      make(map[string]interface{}),  
 evictTick: time.NewTicker(1 * time.Minute),  
 }  
}  
  
func (c *Cache) Set(key string, value interface{}, duration time.Duration) {  
 c.mutex.Lock()  
 defer c.mutex.Unlock()  
  
 // Remove expired keys  
 c.removeExpiredKeys()  
  
 // Check map size limit  
 if c.size() > c.maxBytes {  
 c.evictTick.Stop() // Stop eviction timer if over limit  
 for c.size() > c.maxBytes {  
 // Remove the oldest key to reduce map size  
 delete(c.keys, c.keys[len(c.keys)-1]) // Remove the oldest key  
 }  
 c.evictTick.Reset(1 * time.Minute) // Reset eviction timer  
 }  
  
 // Set new key-value pair with expiration time  
 expiration := time.Now().Add(duration)  
 c.keys[key] = value  
 go func() {  
 time.Sleep(duration) // Wait for expiration duration before evicting key  
 c.Delete(key)         // Remove expired key after duration passed  
 }()  
}  
  
func (c *Cache) Get(key string) interface{} {  
 c.mutex.RLock()  
 defer c.mutex.RUnlock()  
 return c.keys[key]  
}  
  
func (c *Cache) Delete(key string) {  
 c.mutex.Lock()  
 defer c.mutex.Unlock()  
 delete(c.keys, key) // Remove key from cache map  
}  
  
func (c *Cache) removeExpiredKeys() {  
 for key := range c.keys { // Iterate over all keys in the map  
 if time.Now().After(c.keys[key].(time.Time)) { // Check if key has expired  
 delete(c.keys, key) // Remove expired key from cache map  
 }  
 } } 
 func (c *Cache) size() int64 { return int64(len(c.keys)) } 
 func main() { cache := NewCache(100000000) // 100MB cache := NewCache(100000000) // 100MB (100,000,000 bytes) cache := NewCache(1024 * 1024 * 1024) // 1GB cache := NewCache(1024 * 1024 * 1024 * 1024) // 1TB cache := NewCache(5 * 1024 * 1024 * 1024) // 5GB fmt.Println("Initial cache size:", cache.size()) cache.Set("key1", "value1", 5*time.Second) cache.Set("key2", "value2", 15*time.Second) fmt.Println("After setting keys:", cache.size()) time.Sleep(16 * time.Second) fmt.Println("After sleep:", cache.size()) cache.Get("key1") fmt.Println("After get:", cache.size()) time.Sleep(5 * time.Second) fmt.Println("After sleep:", cache.size()) cache.Delete("key1") fmt.Println("After delete:", cache.")```