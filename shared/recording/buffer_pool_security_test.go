package recording

import "testing"

func TestBufferPoolGetHandlesUnexpectedPoolType(t *testing.T) {
	pool := NewBufferPool(16, 16)
	pool.pool.Put("not-a-byte-slice")
	buf := pool.Get()
	if len(buf) != 16*16*3 {
		t.Fatalf("unexpected buffer size: got %d", len(buf))
	}
}

func TestSafeRGB24BufferSizeOverflowFallsBack(t *testing.T) {
	size := safeRGB24BufferSize(maxBufferPoolBytes, maxBufferPoolBytes)
	if size != 1 {
		t.Fatalf("expected overflow fallback size 1, got %d", size)
	}
}
