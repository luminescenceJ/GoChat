package main

import (
	"container/heap"
	"fmt"
	"time"
)

type ListNode struct {
	Key        int
	Val        int
	Next       *ListNode
	Prev       *ListNode
	ExpireTime time.Time
}

type LRU_ttl struct {
	cache map[int]*ListNode
	head  *ListNode
	tail  *ListNode
	size  int
	total int
	ttl   time.Duration
	queue ExpireHeap
}

type ExpireHeap []*ListNode

func (h ExpireHeap) Len() int           { return len(h) }
func (h ExpireHeap) Less(i, j int) bool { return h[i].ExpireTime.Before(h[j].ExpireTime) }
func (h ExpireHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *ExpireHeap) Push(x interface{}) {
	*h = append(*h, x.(*ListNode))
}
func (h *ExpireHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func Init(size int, ttl time.Duration) *LRU_ttl {
	this := &LRU_ttl{}
	this.cache = make(map[int]*ListNode)
	this.head = &ListNode{}
	this.tail = &ListNode{}

	this.head.Next = this.tail
	this.tail.Prev = this.head

	this.ttl = ttl
	this.total = size
	this.queue = make(ExpireHeap, 0)
	heap.Init(&this.queue)
	return this
}

func (this *LRU_ttl) Put(key int, val int) {
	this.Del()
	if v, ok := this.cache[key]; ok {
		// 存储过 ，更新
		v.Val = val
		v.ExpireTime = time.Now().Add(this.ttl)
		this.MoveToHead(v)
		return
	}
	// 没存储过
	if this.total == this.size && this.size > 0 {
		// 删除末尾节点
		toDel := this.tail.Prev
		this.RemoveNode(toDel)
		delete(this.cache, toDel.Key)
		this.size--
	}
	newNode := &ListNode{Key: key, Val: val, ExpireTime: time.Now().Add(this.ttl)}
	this.InsertToHead(newNode)
	this.cache[key] = newNode
	this.size++
	heap.Push(&this.queue, newNode)
}

func (this *LRU_ttl) Del() {
	// 每次访问触发
	for this.queue.Len() > 0 {
		node := this.queue[0]
		if node.ExpireTime.After(time.Now()) {
			break
		}

		heap.Pop(&this.queue)
		this.RemoveNode(node)
		delete(this.cache, node.Key)
		this.size--
	}
}

func (this *LRU_ttl) Get(key int) int {
	this.Del()
	if _, ok := this.cache[key]; !ok {
		return -1
	}
	// key 存在
	v := this.cache[key]
	// 移动到链表头
	this.MoveToHead(v)
	return v.Val
}

func (this *LRU_ttl) InsertToHead(node *ListNode) {
	this.head.Next.Prev = node
	node.Next = this.head.Next
	this.head.Next = node
	node.Prev = this.head
}

func (this *LRU_ttl) RemoveNode(node *ListNode) *ListNode {
	node.Prev.Next = node.Next
	node.Next.Prev = node.Prev
	return node
}

func (this *LRU_ttl) MoveToHead(node *ListNode) {
	node = this.RemoveNode(node)
	this.InsertToHead(node)
	return
}

func lru_test() {
	// 初始化LRU缓存，设置容量为3，TTL为2秒

	cache := Init(3, 2*time.Second)

	fmt.Println("=== 测试1: 基本Put和Get ===")
	cache.Put(1, 100)
	cache.Put(2, 200)
	cache.Put(3, 300)
	fmt.Printf("Get(1): %d (期望: 100)\n", cache.Get(1))
	fmt.Printf("Get(2): %d (期望: 200)\n", cache.Get(2))
	fmt.Printf("Get(3): %d (期望: 300)\n", cache.Get(3))

	fmt.Println("\n=== 测试2: LRU淘汰机制 ===")
	cache.Put(4, 400) // 应该淘汰最久未使用的1
	fmt.Printf("Get(1): %d (期望: -1)\n", cache.Get(1))
	fmt.Printf("Get(4): %d (期望: 400)\n", cache.Get(4))

	fmt.Println("\n=== 测试3: TTL过期机制 ===")
	time.Sleep(3 * time.Second) // 等待所有条目过期
	fmt.Printf("Get(2): %d (期望: -1)\n", cache.Get(2))
	fmt.Printf("Get(3): %d (期望: -1)\n", cache.Get(3))
	fmt.Printf("Get(4): %d (期望: -1)\n", cache.Get(4))

	fmt.Println("\n=== 测试4: 更新已有键值 ===")
	cache.Put(5, 500)
	cache.Put(5, 550) // 更新值
	fmt.Printf("Get(5): %d (期望: 550)\n", cache.Get(5))
	time.Sleep(1 * time.Second)
	cache.Get(5)
	time.Sleep(2 * time.Second) // 等待初始TTL过期
	fmt.Printf("Get(5) after renewal: %d (期望: -1)\n", cache.Get(5))
}
