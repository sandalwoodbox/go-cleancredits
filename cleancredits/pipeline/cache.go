package pipeline

import (
	"fmt"
	"image"
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
	"gocv.io/x/gocv"
)

// FrameCache allows caching recently loaded frames and also controls locking
// for the VideoCapture to avoid contention between threads.
type FrameCache struct {
	vc     *gocv.VideoCapture
	locker *sync.Mutex
	cache  *lru.Cache[int, image.Image]
}

func NewFrameCache(vc *gocv.VideoCapture) (*FrameCache, error) {
	cache, err := lru.New[int, image.Image](128)
	if err != nil {
		return nil, fmt.Errorf("creating cache: %v", err)
	}
	return &FrameCache{
		vc:     vc,
		locker: &sync.Mutex{},
		cache:  cache,
	}, nil
}

func (fc *FrameCache) LoadFrame(n int) (image.Image, error) {
	fmt.Println("Loading frame: ", n)
	img, ok := fc.cache.Get(n)
	if !ok {
		fmt.Println("Cache miss for frame: ", n)
		m := gocv.NewMat()
		defer m.Close()
		fc.locker.Lock()
		fc.vc.Set(
			gocv.VideoCapturePosFrames,
			float64(n),
		)
		ok := fc.vc.Read(&m)
		fc.locker.Unlock()
		if !ok {
			return nil, fmt.Errorf("invalid frame number: %d", n)
		}
		var err error
		img, err = m.ToImage()
		if err != nil {
			return nil, fmt.Errorf("converting to image: %v", err)
		}
		fc.cache.Add(n, img)
	}
	return img, nil
}
