package main

import (
	"time"
	gosifter "github.com/jtuki/gosifter/api"
	bmi "github.com/jtuki/gosifter/benchmark/internal"
	"strconv"
	"fmt"
	"sync"
	"runtime"
	"encoding/json"
	"sync/atomic"
	"flag"
)

const (
	BENCHMARK_RUN_DURATION = 1*time.Minute // several minutes
	BENCHMARK_STAT_TIMETICK = 5*time.Second
	BENCHMARK_DATA_SET_LENGTH = 100000
)

var (
	gl_deviceInfoList []*bmi.DeviceInfo
	gl_runtime_cpu_num int
)

func prepare_benchmark_data_set() {
	t0 := time.Now()
	fmt.Printf("[%s]  start: prepare_benchmark_data_set\n", t0.Format("2006-01-02 15:04:05.000-0700"))
	
	gl_deviceInfoList = make([]*bmi.DeviceInfo, BENCHMARK_DATA_SET_LENGTH)
	for i := 0; i < BENCHMARK_DATA_SET_LENGTH; i++ {
		gl_deviceInfoList[i] = &bmi.DeviceInfo{
			Extended: bmi.DeviceInfoExtended{
				ImageUrl: "http://image.com/image_uri/" + strconv.Itoa(i),
			},
			Meta: bmi.DeviceInfoMeta{
				Country: "us+china+eu+south-africa",
				Province: "whatever-province",
				City: "some-city",
				IP: "1.2.3.4",
				ModVersion: "mod-version-1-2-3" + strconv.Itoa(i),
				DevVersion: "dev-version-4-5-6" + strconv.Itoa(i),
			},
		}
		gl_deviceInfoList[i].DeviceInfoBasic = bmi.DeviceInfoBasic{
			Domain: i*2,
			SubDomain: i*10,
			PhysicalDeviceId: "123456ABCDEF" + strconv.Itoa(i),
		}
	}
	
	t1 := time.Now()
	fmt.Printf("[%s] finish: prepare_benchmark_data_set; duration[%f(s)]\n", t1.Format("2006-01-02 15:04:05.000-0700"), t1.Sub(t0).Seconds())
}

type marshal_func func(v interface{}) ([]byte, error)

func marshal_routine(routinesNum int, mf marshal_func) {
	var wg sync.WaitGroup

	// 运行超时
	timeout := time.NewTimer(BENCHMARK_RUN_DURATION)
	defer timeout.Reset(0)
	
	// 定时器ticker（周期输出统计结果）
	ticker := time.Tick(BENCHMARK_STAT_TIMETICK)

	totalBytes := int64(0) // 总共序列化结果的字节数（简单的模拟对序列化的处理）
	totalTask := int64(0) // 总共序列化的个数
	
	// 是否已经结束运行（0/1: 运行进行中/已结束）
	var stopped int32
	
	// 起始时间
	t0 := time.Now()

	// 输出统计信息
	perform_stat_func := func(now time.Time, begin time.Time) {
		tb, tt := atomic.LoadInt64(&totalBytes), atomic.LoadInt64(&totalTask)
		fmt.Printf("[%s][%5.2f] total_tasks[%20d] total_bytes[%20d]\n",
			now.Format("2006-01-02 15:04:05.000-0700"), now.Sub(begin).Seconds(), tt, tb)
	}
	
	for i := 0; i < routinesNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			idx := 0
			for {
				if atomic.LoadInt32(&stopped) == 1 {
					break
				}
				
				for ; idx < BENCHMARK_DATA_SET_LENGTH; idx++ {
					marshalBytes, err := mf(gl_deviceInfoList[idx])
					if err != nil {
						// 直接报错即可
						panic(err)
					} else {
						atomic.AddInt64(&totalBytes, int64(len(marshalBytes)))
						atomic.AddInt64(&totalTask, 1)
						
						// 给检查 stopped 一些时间
						if idx % 100 == 0 {
							break
						}
					}
				}
				
				idx %= BENCHMARK_DATA_SET_LENGTH
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			
			for {
				select {
				case <-ticker:
					perform_stat_func(time.Now(), t0)
				case <-timeout.C:
					atomic.StoreInt32(&stopped, 1)
					n := time.Now()
					time.Sleep(time.Second)
					perform_stat_func(n, t0)
				}
			}
		}()
	}
	
	wg.Wait()
}

func gosifter_marshal(v interface{}) ([]byte, error) {
	m, err := gosifter.SiftStruct(v, gosifter.CONFIDENTIAL_LEVEL_MAX)
	if err != nil {
		return nil, err
	}
	return json.Marshal(m)
}

func main() {
	gl_runtime_cpu_num = runtime.NumCPU()
	runtime.GOMAXPROCS(gl_runtime_cpu_num)

	marshalType := flag.Int("type", 1, "type of marshaller to use: json[1], gosifter[2]")
	
	switch *marshalType {
	case 1:
		prepare_benchmark_data_set()
		marshal_routine(gl_runtime_cpu_num, marshal_func(json.Marshal))
		return
	case 2:
		prepare_benchmark_data_set()
		marshal_routine(gl_runtime_cpu_num, marshal_func(gosifter_marshal))
		return
	default:
		fmt.Println("unsupported marshalType: json[1], gosifter[2]")
		return
	}
}