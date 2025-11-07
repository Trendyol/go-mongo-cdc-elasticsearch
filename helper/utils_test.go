package helper

import (
	"testing"
)

func TestChunkSlice(t *testing.T) {
	size := 1009
	slice := make([]int, size)
	for i := 0; i < size; i++ {
		slice[i] = 0
	}

	chunks := ChunkSlice[int](slice, 6)

	if len(chunks) != 6 {
		t.Errorf("ChunkSliceWithSize failed")
	}

	for idx, chunk := range chunks {
		if idx+1 == 1 && len(chunk) == 169 {
			continue
		}

		if len(chunk) == 168 {
			continue
		}

		t.Errorf("ChunkSliceWithSize failed")
	}
}
