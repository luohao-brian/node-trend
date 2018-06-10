package main

import (
	"container/ring"
	"fmt"
	linuxproc "github.com/c9s/goprocinfo/linux"
	"github.com/sajari/regression"
	"time"
)

func getCPUSample() (idle, total uint64) {
	stat, err := linuxproc.ReadStat("/proc/stat")
	if err != nil {
		fmt.Errorf("stat read fail")
	}

	idle = stat.CPUStatAll.Idle
	total = stat.CPUStatAll.Idle +
		stat.CPUStatAll.User +
		stat.CPUStatAll.System +
		stat.CPUStatAll.Nice +
		stat.CPUStatAll.IOWait +
		stat.CPUStatAll.IRQ +
		stat.CPUStatAll.SoftIRQ +
		stat.CPUStatAll.Steal +
		stat.CPUStatAll.Guest +
		stat.CPUStatAll.GuestNice

	return idle, total
}

func getMemSample() (avail, total uint64) {
	info, err := linuxproc.ReadMemInfo("/proc/meminfo")
	if err != nil {
		fmt.Errorf("meminfo read fail")
	}

	return info.MemAvailable, info.MemTotal
}

func main() {
	cpuList := ring.New(60)
	memList := ring.New(60)

	for {
		cpuIdle0, cpuTotal0 := getCPUSample()
		time.Sleep(1 * time.Second)
		cpuIdle1, cpuTotal1 := getCPUSample()
		cpuIdleTicks := float64(cpuIdle1 - cpuIdle0)
		cpuTotalTicks := float64(cpuTotal1 - cpuTotal0)
		cpuUsage := 100 * (cpuTotalTicks - cpuIdleTicks) / cpuTotalTicks
		fmt.Printf("CPU usage is %f%%\n", cpuUsage)

		memAvail, memTotal := getMemSample()
		memUsage := 100 * (1.0 - float64(memAvail)/float64(memTotal))
		fmt.Printf("Memory usage is %f%%\n", memUsage)

		cpuList.Value = cpuUsage
		cpuList = cpuList.Next()
		memList.Value = memUsage
		memList = memList.Next()

		cpuRegression := new(regression.Regression)
		prevCpuX := float64(0.0)
		cpuList.Do(func(x interface{}) {
			if x != nil {
				//fmt.Printf("Item: %v %v\n", x, prevCpuX)
				cpuRegression.Train(regression.DataPoint(x.(float64), []float64{prevCpuX}))
				prevCpuX = x.(float64)
			}
		})
		cpuRegression.Run()
		fmt.Printf("CPU usage regression formula:\n%v\n", cpuRegression.Formula)

		memRegression := new(regression.Regression)
		prevMemX := float64(0.0)

		memList.Do(func(x interface{}) {
			if x != nil {
				//fmt.Printf("Item: %v %v\n", x, prevMemX)
				memRegression.Train(regression.DataPoint(x.(float64), []float64{prevMemX}))
				prevMemX = x.(float64)
			}
		})
		memRegression.Run()
		fmt.Printf("Memory usage regression formula:\n%v\n", memRegression.Formula)
	}
}
