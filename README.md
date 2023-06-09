# LocalCache
一个golang的支持LRU的进程内缓存库，规避了map占用的内存不归还给操作系统的问题。

localcache内部使用了一个map和一个双链表来保存数据。map作为主要的存储数据结果。
而双链表则用来实现O(1)的lru淘汰策略

随着运行时间的增长，map中被淘汰的key也会逐渐增加。但是go的map并不能保证将已经使用内存空间释放并归还给操作系统。
所以内部维护了一个计数器，当淘汰的key超过半数时，通过捯换的方式将key放入一个新的map从而保证map占用的内存空间可以被释放。

## Usage

```
package main

import (
    "fmt"
    "time"
    "github.com/nobugtodebug/localcache"
)

func main() {
    lc := localcache.NewCache[string, int](localcache.OptWithExpire(time.Millisecond * 200))
    n := 1
    lc.Set("a", &n)

    v, ok := lc.Get("a")
    if ok {
        fmt.Println("key:", "a", "value:", *v)
    }

    time.Sleep(time.Second)
    v, ok = lc.Get("a")
    if !ok {
        fmt.Println("key:", "a", "is expired")
    }
}
```