package main

import (
    "io/ioutil"
	"os"
	"os/signal"
	"encoding/binary"
	"syscall"
	"strconv"
	"time"
	"github.com/sevlyar/go-daemon"
	"log"
)

func check(e error) {
    if e != nil {
        panic(e)
    }
}

type keyEvent struct {
	Sec int64
	Usec int64
	EvtType uint16
	Code uint16
	Value int32
}

func main() {
	cntxt := &daemon.Context{
		PidFileName: "pid",
		PidFilePerm: 0644,
		LogFileName: "log",
		LogFilePerm: 0660,
		WorkDir:     "./",
		Umask:       027,
		Args:        []string{"[go-daemon sample]"},
	}

	d, err := cntxt.Reborn()

	if err != nil {
		log.Fatal("Unable to run: ", err)
	}

	if d != nil {
		return
	}

	defer cntxt.Release()

	log.Print("daemonized successfully");

	callstring := []byte("\\_SB.ALS._ALR")
	err = ioutil.WriteFile("/proc/acpi/call", callstring, 0000)
	check(err)

	dat, err := ioutil.ReadFile("/proc/acpi/call")
	check(err)

	//fmt.Println(string(dat))

//	err = syscall.Setuid(syscall.Getuid())
//	check(err)
//	log.Print("dropped priveleges")
	
	alsVals, blAdjs := parsePointString(string(dat))

	log.Printf("%v, %v", alsVals, blAdjs)

	evtHandle, err := os.Open("/dev/input/event7")
	check(err)

	susp_resume := make(chan int)

	log.Println("calling adjbright")

	go adjBright(susp_resume, alsVals, blAdjs)

	var evt keyEvent
	for {
		check(binary.Read(evtHandle, binary.LittleEndian, &evt))
		if evt.EvtType == 1 /*KEY*/ &&
			evt.Code == 0x230 /*KEY_ALS_TOGGLE*/ &&
			evt.Value == 1 {
			susp_resume <- 1
		}
	}
}

func adjBright(susp_resume chan int, alsVals []int64, blAdjs []int64) {
	curbrightStr, err := ioutil.ReadFile("/sys/class/backlight/intel_backlight/brightness")
	check(err)
	curbright, err := strconv.Atoi(string(curbrightStr[:len(curbrightStr)-1]))
	check(err)
	
	baseBright, prevBright := curbright, curbright
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func(){
		for _ = range c {
			err = ioutil.WriteFile("/sys/class/backlight/intel_backlight/brightness",
				[]byte(strconv.Itoa(baseBright)),
				0000)
			check(err)
			os.Exit(0)
		}
	}()
	for {
		curbrightStr, err = ioutil.ReadFile("/sys/class/backlight/intel_backlight/brightness")
		check(err)
		curbright, err = strconv.Atoi(string(curbrightStr[:len(curbrightStr)-1]))
		check(err)
		
		if curbright != prevBright {
			baseBright += curbright - prevBright
		}
		

		illumString, err := ioutil.ReadFile("/sys/bus/acpi/devices/ACPI0008:00/iio:device0/in_illuminance_input")
		check(err)

		illum_i, err := strconv.Atoi(string(illumString[:len(illumString)-1]))
		check(err)

		illum := int64(illum_i)

		var illumApprox int
		for i := range alsVals {
			if alsVals[i] <= illum && alsVals[i+1] > illum {
				illumApprox = i
			}
		}

		log.Print(illumApprox)

		brightAdj := float64(calcBrightAdj(illum,
			alsVals[illumApprox:illumApprox+2],
			blAdjs[illumApprox:illumApprox+2])) / 100.0

		brightVal := int(brightAdj * float64(baseBright))
		prevBright = brightVal

		incTicker := time.NewTicker(time.Millisecond * 1)

		var sign int
		if brightVal - curbright < 0 { sign = -1 } else { sign = 1 }
		for tempBr := curbright; tempBr != brightVal; tempBr+=sign {
			err = ioutil.WriteFile("/sys/class/backlight/intel_backlight/brightness",
				[]byte(strconv.Itoa(tempBr)),
				0000)
			if err != nil {
				log.Printf("Warning: Writing brightness: %v", err)
			}
			<-incTicker.C
		}
		incTicker.Stop()
		
		log.Printf("Wrote brightness %v", brightVal)
		select {
		case <-susp_resume:
			<-susp_resume
		case <-time.After(time.Millisecond * 100):
			
		}
	}
}
