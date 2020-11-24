package redisactives

import (
	"time"

	"github.com/herb-go/herbdata"
)

func timekey(t int64, d time.Duration, prev bool) []byte {
	buf := make([]byte, 8)
	tindex := uint64((t * int64(time.Second) / int64(d)))
	if prev {
		tindex = tindex - 1
	}
	herbdata.DataOrder.PutUint64(buf, tindex)
	return buf
}
