package book

import (
	"math/rand"
	"time"
)

// maxLevel 为跳表最大层数（2^16 ≈ 6.5 万个价格档位，对订单簿足够）。
const maxLevel = 16

type node struct {
	key  int64
	next []*node
}

// Skiplist 是按 key 升序排列的跳表，保存价格档位集合。
//
// 为什么用跳表而不是 std::map / 红黑树：
//   - 撮合引擎只需要按价格有序取最值（买盘最大、卖盘最小）与增删档位；
//   - 跳表在插入/删除/取最值上都是 O(log n)，实现简单、便于并发改造
//     （LevelDB / Redis 的 zset 均采用同类结构）；
//   - 没有外部依赖，单文件即可讲清全部实现。
type Skiplist struct {
	head  *node
	level int
	rnd   *rand.Rand
}

// NewSkiplist 创建一个空跳表。
func NewSkiplist() *Skiplist {
	return &Skiplist{
		head:  &node{next: make([]*node, maxLevel)},
		level: 1,
		rnd:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// randomLevel 按 1/2 概率逐层上升，生成新节点的层数。
func (s *Skiplist) randomLevel() int {
	lv := 1
	for lv < maxLevel && s.rnd.Intn(2) == 0 {
		lv++
	}
	return lv
}

// Insert 插入 key，返回是否为新 key（已存在则返回 false，不重复插入）。
func (s *Skiplist) Insert(key int64) bool {
	update := make([]*node, maxLevel)
	cur := s.head
	for i := s.level - 1; i >= 0; i-- {
		for cur.next[i] != nil && cur.next[i].key < key {
			cur = cur.next[i]
		}
		update[i] = cur
	}
	if nxt := cur.next[0]; nxt != nil && nxt.key == key {
		return false
	}
	lv := s.randomLevel()
	if lv > s.level {
		for i := s.level; i < lv; i++ {
			update[i] = s.head
		}
		s.level = lv
	}
	nd := &node{key: key, next: make([]*node, lv)}
	for i := 0; i < lv; i++ {
		nd.next[i] = update[i].next[i]
		update[i].next[i] = nd
	}
	return true
}

// Delete 删除 key，返回是否存在。
func (s *Skiplist) Delete(key int64) bool {
	update := make([]*node, maxLevel)
	cur := s.head
	for i := s.level - 1; i >= 0; i-- {
		for cur.next[i] != nil && cur.next[i].key < key {
			cur = cur.next[i]
		}
		update[i] = cur
	}
	tgt := cur.next[0]
	if tgt == nil || tgt.key != key {
		return false
	}
	for i := 0; i < s.level; i++ {
		if update[i].next[i] != tgt {
			break
		}
		update[i].next[i] = tgt.next[i]
	}
	for s.level > 1 && s.head.next[s.level-1] == nil {
		s.level--
	}
	return true
}

// Min 返回最小 key 及是否存在。
func (s *Skiplist) Min() (int64, bool) {
	if s.head.next[0] == nil {
		return 0, false
	}
	return s.head.next[0].key, true
}

// Max 返回最大 key 及是否存在。
func (s *Skiplist) Max() (int64, bool) {
	cur := s.head
	for i := s.level - 1; i >= 0; i-- {
		for cur.next[i] != nil {
			cur = cur.next[i]
		}
	}
	if cur == s.head {
		return 0, false
	}
	return cur.key, true
}

// Ascend 按升序遍历所有 key，直到 fn 返回 false。
func (s *Skiplist) Ascend(fn func(key int64) bool) {
	for cur := s.head.next[0]; cur != nil; cur = cur.next[0] {
		if !fn(cur.key) {
			return
		}
	}
}
