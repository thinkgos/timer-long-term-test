package main

import (
	"log"
	"math"
	"math/rand/v2"
	"net/http"
	"sync/atomic"
	"time"

	_ "net/http/pprof"

	"github.com/thinkgos/logger"
	"github.com/thinkgos/timer"
)

// almost 1,000,000 task
func main() {
	logger.ReplaceGlobals(logger.NewLogger(
		logger.WithAdapter(logger.AdapterFile),
		logger.WithFilename("./monitor-long-term.log"),
		logger.WithEnableLocalTime(),
		logger.WithEnableCompress(),
	))

	go func() {
		sum := &atomic.Int64{}
		t := time.NewTicker(time.Second)
		for {
			<-t.C
			added := 0
			ranv := rand.IntN(10)
			max := int(rand.Uint32N(math.MaxUint16 << 2))
			for i := 100; i < max; i += 200 {
				added++
				ii := i + ranv

				timer.Go(func() {
					sum.Add(1)
					delayms := int64(ii) * 20
					task := timer.NewTask(time.Duration(delayms) * time.Millisecond).WithJob(&job{
						sum:          sum,
						expirationMs: time.Now().UnixMilli() + delayms,
					})
					timer.AddTask(task)

					// for test race
					// if ii%0x03 == 0x00 {
					// 	timer.Go(func() {
					// 		task.Cancel()
					// 	})
					// }
				})
			}
			log.Printf("task: %v - %v added: %d", timer.TaskCounter(), sum.Load(), added)
		}
	}()

	addr := ":9990"
	log.Printf("http stated '%v'\n", addr)
	log.Println(http.ListenAndServe(addr, nil))
}

type job struct {
	sum          *atomic.Int64
	expirationMs int64
}

func (j *job) Run() {
	j.sum.Add(-1)
	now := time.Now().UnixMilli()
	diff := now - j.expirationMs
	if diff > 1 {
		log.Printf("this task no equal, diff: %d %d %d\n", now, j.expirationMs, diff)
	}
	switch {
	case diff > 16000:
		logger.OnError().
			Int64("now", now).
			Int64("expirationAt", j.expirationMs).
			Int64("diff", diff).
			Msg("this task large diff")
	case diff > 100:
		logger.OnWarn().
			Int64("now", now).
			Int64("expirationAt", j.expirationMs).
			Int64("diff", diff).
			Msg("this task large diff")

	}

}
