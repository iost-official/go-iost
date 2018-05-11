package prque

const blockSize = 4096

type item struct {
	value interface{}
	priority float32
}

type sstack struct {
	size int
	capacity int
	offset int

	blocks [][]*item
	active []*item
}

func newSstack() *sstack {
	result := new(sstack)
	result.active = make([]*item, blockSize)
	result.blocks = [][]*item{result.active}
	result.capacity = blockSize
	return result
}

func (s *sstack) Push(data interface{}) {
	if s.size == s.capacity {
		s.active = make([]*item, blockSize)
		s.blocks = append(s.blocks, s.active)
		s.capacity += blockSize
		s.offset = 0
	} else if s.offset == blockSize {
		s.active = s.blocks[s.size / blockSize]
		s.offset = 0
	}
	s.active[s.offset] = data.(*item)
	s.offset++
	s.size++
}

