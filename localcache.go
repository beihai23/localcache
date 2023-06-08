package localcache

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

type LCache[K comparable, V any] struct {
	kvStore    map[K]*lruNode[K, V] // 保存数据的hashmap，提供O(1)的查找能力
	lruHead    *lruNode[K, V]       // lru链表的表头指针
	lruTail    *lruNode[K, V]       // lru链表的表尾指针
	lock       sync.RWMutex         // 保护map的锁
	ch         chan *lruNode[K, V]  // 异步更新lru链表
	o          CacheOptions
	keyCounter int
}

type lruNode[K comparable, V any] struct {
	k      K
	v      *V
	exp    time.Duration
	expAt  time.Time
	next   *lruNode[K, V]
	prev   *lruNode[K, V]
	rmFlag bool
}

// CacheOptions 本地的缓存选项
type CacheOptions struct {
	exp       time.Duration // 默认的过期时间
	max       int           // 缓存的key数量上限
	maxMemory int           // 缓存的内存上限
}

type Option func(co *CacheOptions)

// OptWithExpire 设置默认的过期时间
func OptWithExpire(exp time.Duration) Option {
	return func(co *CacheOptions) {
		co.exp = exp
	}
}

// OptWithMaxKeys 设置缓存的key数量上限
func OptWithMaxKeys(max int) Option {
	return func(co *CacheOptions) {
		co.max = max
	}
}

// OptWithMaxMemory 设置缓存的内存上限
func OptWithMaxMemory(maxMemory string) Option {
	maxMemory = strings.ToUpper(maxMemory)
	return func(co *CacheOptions) {
		if strings.HasSuffix(maxMemory, "GB") {
			maxMemory = strings.TrimSuffix(maxMemory, "GB")
			n, _ := strconv.Atoi(maxMemory)
			co.maxMemory = n * 1024 * 1024 * 1024
		} else if strings.HasSuffix(maxMemory, "MB") {
			maxMemory = strings.TrimSuffix(maxMemory, "MB")
			n, _ := strconv.Atoi(maxMemory)
			co.maxMemory = n * 1024 * 1024
		} else if strings.HasSuffix(maxMemory, "KB") {
			maxMemory = strings.TrimSuffix(maxMemory, "KB")
			n, _ := strconv.Atoi(maxMemory)
			co.maxMemory = n * 1024
		}
	}
}

func NewCache[K comparable, V any](opts ...Option) *LCache[K, V] {
	o := &CacheOptions{}
	for _, opt := range opts {
		opt(o)
	}

	lc := &LCache[K, V]{}
	lc.o = *o
	lc.kvStore = make(map[K]*lruNode[K, V])
	lc.ch = make(chan *lruNode[K, V], 5)
	lc.lruHead = &lruNode[K, V]{}
	lc.lruTail = &lruNode[K, V]{}
	lc.lruHead.next = lc.lruTail
	lc.lruTail.prev = lc.lruHead

	go lc.asyncJob()

	return lc
}

// Set 设置/更新缓存内容
func (lc *LCache[K, V]) Set(key K, value *V) {
	lc.lock.Lock()
	defer lc.lock.Unlock()

	n, ok := lc.kvStore[key]
	if !ok {
		n = &lruNode[K, V]{
			k:   key,
			v:   value,
			exp: lc.o.exp,
		}
		lc.keyCounter += 1 // 累加map历史上保存过多少个key
	}
	n.v = value

	lc.kvStore[key] = n

	// 刷新缓存时间
	lc.ch <- n
}

// Get 读取缓存内容
func (lc *LCache[K, V]) Get(key K) (value *V, ok bool) {
	lc.lock.RLock()
	defer lc.lock.RUnlock()

	n, ok := lc.kvStore[key]
	if !ok {
		return nil, false
	}

	// 刷新缓存时间
	lc.ch <- n

	return n.v, true
}

// Del 读取缓存内容
func (lc *LCache[K, V]) Del(key K) {
	lc.lock.Lock()
	defer lc.lock.Unlock()

	n, ok := lc.kvStore[key]
	if !ok {
		return
	}
	delete(lc.kvStore, key)
	n.rmFlag = true

	// 刷新缓存时间
	lc.ch <- n
}

//----

// asyncJob 处理lru的更新，以及定时清理过期的缓存内容
func (lc *LCache[K, V]) asyncJob() {
	t := time.NewTicker(time.Millisecond * 50)
	for {
		select {
		case n, ok := <-lc.ch:
			if !ok {
				break
			}

			// 更新过期时间
			n.expAt = time.Now().Add(n.exp)

			if n.prev != nil && n.next != nil {
				// 将n从链表中摘除
				n.prev.next = n.next
				n.next.prev = n.prev
				n.prev = nil
				n.next = nil
			}

			if !n.rmFlag {
				// 将n插入表头
				n.prev = lc.lruHead
				n.next = lc.lruHead.next
				lc.lruHead.next.prev = n
				lc.lruHead.next = n
			}
		case <-t.C:
			// 清理已过期的值
			now := time.Now()

			// 从尾部向前遍历
			for n := lc.lruTail.prev; n != lc.lruHead; n = n.prev {
				if now.After(n.expAt) {
					fmt.Println(n.k, "expired")
					// 将n从链表中摘除
					n.prev.next = n.next
					n.next.prev = n.prev

					lc.lock.Lock()
					delete(lc.kvStore, n.k)
					lc.lock.Unlock()
				} else {
					// 当所有k的过期时间一致时，可以直接结束
					break
				}
			}

			// map中当前的key数量只有历史上的一半时，就清理一次map
			if len(lc.kvStore) < lc.keyCounter/2 {
				// 将当前map中的内容转移到新的map中
				newMap := make(map[K]*lruNode[K, V])
				lc.lock.RLock()
				for k, v := range lc.kvStore {
					newMap[k] = v
				}
				lc.lock.RUnlock()

				// 替换掉老的map
				lc.lock.Lock()
				lc.kvStore = newMap
				lc.keyCounter = len(lc.kvStore)
				lc.lock.Unlock()
			}
		}
	}
}

func (lc *LCache[K, V]) dumpLink() {
	fmt.Println("dumpLink:")
	// 从尾部向前遍历
	for n := lc.lruTail; n != nil; n = n.prev {
		if n.next == nil {
			fmt.Printf("tail %p\n", &*n)
		} else if n.prev == nil {
			fmt.Printf("head %p\n", &*n)
		} else {
			fmt.Println("node", "key", n.k, "next", &*n.next, "prev", &*n.prev)
		}
	}
}
