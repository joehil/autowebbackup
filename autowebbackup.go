package main

import (
	"log"
	"io/ioutil"
	"os"
	"os/signal"
	"os/exec"
	"fmt"
	"strings"
	"strconv"
	"syscall"
	"hash/fnv"
	"github.com/illarion/gonotify"
	"github.com/spf13/viper"
	"github.com/natefinch/lumberjack"
)

//var read_log1 string = "/var/log/monit.log"
//var read_log2 string = "/var/log/virtualmin/remote-browser.eu_error_log"
var do_trace bool = false
var msg_trace bool = false
var pidfile string
var ownlog string
var logs []string
var rlogs []*os.File
var rpos []int64
var loghash []uint32


func main() {
// Set location of config 
	viper.SetConfigName("proc_logs") // name of config file (without extension)
	viper.AddConfigPath("/etc/")   // path to look for the config file in

// Read config
	read_config()

// Get commandline args
	if len(os.Args) > 1 {
        	a1 := os.Args[1]
        	if a1 == "reload" {
			b, err := ioutil.ReadFile(pidfile) 
    			if err != nil {
        			log.Fatal(err)
    			}
			s := string(b)
			fmt.Println("Reload", s)
			cmd := exec.Command("kill", "-hup", s)
                	_ = cmd.Start()
                	os.Exit(0)
        	}
                if a1 == "mtraceon" {
                        b, err := ioutil.ReadFile(pidfile)
                        if err != nil {
                                log.Fatal(err)
                        }
                        s := string(b)
                        fmt.Println("MsgTraceOn")
                        cmd := exec.Command("kill", "-10", s)
                        _ = cmd.Start()
                        os.Exit(0)
                }
                if a1 == "mtraceoff" {
                        b, err := ioutil.ReadFile(pidfile)
                        if err != nil {
                                log.Fatal(err)
                        }
                        s := string(b)
                        fmt.Println("MsgTraceOff")
                        cmd := exec.Command("kill", "-12", s)
                        _ = cmd.Start()
                        os.Exit(0)
                }
                if a1 == "run" {
                        proc_run()
                }
		fmt.Println("parameter invalid")
		os.Exit(-1)
	}
	if len(os.Args) == 1 {
		myUsage()
	}
}

func read_config() {
        err := viper.ReadInConfig() // Find and read the config file
        if err != nil { // Handle errors reading the config file
                log.Fatalf("Config file not found: %v", err)
        }

        pidfile = viper.GetString("pid_file")
        if pidfile =="" { // Handle errors reading the config file
                log.Fatalf("Filename for pidfile unknown: %v", err)
        }
        ownlog = viper.GetString("own_log")
        if ownlog =="" { // Handle errors reading the config file
                log.Fatalf("Filename for ownlog unknown: %v", err)
        }
	logs = viper.GetStringSlice("logs")
        do_trace = viper.GetBool("do_trace")

	if do_trace {
		log.Println("do_trace: ",do_trace)
		log.Println("own_log; ",ownlog)
		log.Println("pid_file: ",pidfile)
		for i, v := range logs {
			log.Printf("Index: %d, Value: %v\n", i, v )
		}
	}
}

func myUsage() {
     fmt.Printf("Usage: %s argument\n", os.Args[0])
     fmt.Println("Arguments:")
     fmt.Println("run           Run progam as daemon")
     fmt.Println("reload        Make running daemon reload it's configuration")
     fmt.Println("mtraceon      Make running daemon switch it's message tracing on (useful for coding new rules)")
     fmt.Println("mtraceoff     Make running daemon switch it's message tracing off")
}
