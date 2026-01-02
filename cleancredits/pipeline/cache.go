package pipeline

import (
	"fmt"
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
	"gocv.io/x/gocv"
)

// FrameCache allows caching recently loaded frames and also controls locking
// for the VideoCapture to avoid contention between threads.
type FrameCache struct {
	vc     *gocv.VideoCapture
	locker *sync.Mutex
	cache  *lru.Cache[int, gocv.Mat]
	debug  bool
}

func NewFrameCache(vc *gocv.VideoCapture, debug bool) (*FrameCache, error) {
	cache, err := lru.NewWithEvict(10, func(k int, v gocv.Mat) {
		if debug {
			fmt.Printf("Evicted frame %d. Ptr: %v", k, v.Ptr())
		}
		v.Close()
	})
	if err != nil {
		return nil, fmt.Errorf("creating cache: %v", err)
	}
	return &FrameCache{
		vc:     vc,
		locker: &sync.Mutex{},
		cache:  cache,
		debug:  debug,
	}, nil
}

func (fc *FrameCache) LoadFrame(n int) (gocv.Mat, error) {
	fc.locker.Lock()
	mat, ok := fc.cache.Get(n)
	if !ok {
		mat = gocv.NewMat()
		fc.vc.Set(
			gocv.VideoCapturePosFrames,
			float64(n),
		)
		ok := fc.vc.Read(&mat)
		if !ok {
			return gocv.NewMat(), fmt.Errorf("invalid frame number: %d", n)
		}
		fc.cache.Add(n, mat)
		if fc.debug {
			fmt.Printf("Added frame %d. Ptr: %v\n", n, mat.Ptr())
		}
	} else if fc.debug {
		fmt.Printf("Loaded frame %d. Ptr: %v\n", n, mat.Ptr())
	}
	fc.locker.Unlock()
	return mat, nil
}
