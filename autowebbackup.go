package main

import (
	"log"
	"os"
	"os/exec"
	"fmt"
	"io"
	"time"
	"strings"
	"github.com/spf13/viper"
	"github.com/natefinch/lumberjack"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

var do_trace bool = true

var ownlog string

var dirs []string

var tarcmd string

var ftpsuser string
var ftpspassword string
var ftpshost string

var dailydir string
var weeklydir string
var monthlydir string
var tempdir string

var dailykeep int64
var weeklykeep int64
var monthlykeep int64

var do_encrypt bool = true
var encryptsuffix string
var encryptpassw string

var transferfile string = "/autowebbackup.tar.gz"
var transfersuffix string = "tar.gz"

var ownlogger io.Writer

var hostKey ssh.PublicKey

func main() {
// Set location of config 
	viper.SetConfigName("autowebbackup") // name of config file (without extension)
	viper.AddConfigPath("/etc/")   // path to look for the config file in

// Read config
	read_config()

// Get commandline args
	if len(os.Args) > 1 {
        	a1 := os.Args[1]
        	if a1 == "backup" {
			backup()
			os.Exit(0)
        	}
                if a1 == "list" {
                        list()
			os.Exit(0)
                }
                if a1 == "fetch" {
                        fetch(os.Args[2])
			os.Exit(0)
                }
                if a1 == "decrypt" {
                        decrypt()
                        os.Exit(0)
                }
		fmt.Println("parameter invalid")
		os.Exit(-1)
	}
	if len(os.Args) == 1 {
		myUsage()
	}
}


func backup() {
t := time.Now()
tstr := t.Format("20060102")
wd := t.Weekday()
fstr := tstr[0:6] + "01"
tunix := t.Unix()
daynum := t.Day()

buffer := make([]byte, 8192)

//var ftplogger io.Writer = nil

//if do_trace {
//	ftplogger = ownlogger
//}

config := &ssh.ClientConfig{
	User: ftpsuser,
	Auth: []ssh.AuthMethod{
		ssh.Password(ftpspassword),
	},
	HostKeyCallback: ssh.InsecureIgnoreHostKey(),
}

fmt.Println("Backup started")

conn, err := ssh.Dial("tcp", ftpshost+":22", config)
if err != nil {
	fmt.Println("Failed to dial: ", err)
	log.Fatal("Failed to dial: ", err)
}

// open an SFTP session over an existing ssh connection.
client, err := sftp.NewClient(conn)
if err != nil {
	fmt.Println(err)
	log.Fatal(err)
}

// walk a directory
w := client.Walk(dailydir)
for w.Step() {
	if w.Err() != nil {
		continue
	}
	fmt.Println(w.Path())
        if w.Stat().ModTime().Unix() < tunix - 86400 * dailykeep {
                log.Println("Delete file ",w.Path())
                client.Remove(w.Path())
        }
}

w = client.Walk(weeklydir)
for w.Step() {
        if w.Err() != nil {
                continue
        }
        fmt.Println(w.Path())
        if w.Stat().ModTime().Unix() < tunix - 86400 * dailykeep {
                log.Println("Delete file ",w.Path())
                client.Remove(w.Path())
        }
}

w = client.Walk(monthlydir)
for w.Step() {
        if w.Err() != nil {
                continue
        }
        fmt.Println(w.Path())
        if w.Stat().ModTime().Unix() < tunix - 86400 * dailykeep {
                log.Println("Delete file ",w.Path())
                client.Remove(w.Path())
        }
}

client.Close()
conn.Close()


if do_encrypt {
	transferfile = "/autowebbackup." + encryptsuffix
	transfersuffix = encryptsuffix
}

// Loop over directories
	for i, s := range dirs {
		var cmd *exec.Cmd
		var parm string
    		fmt.Println(i, s)
		if len(os.Args) > 2 {
			parm = os.Args[2]
		} else {
			parm = os.Args[1]
		}
                if daynum == 1 || parm == "full" {
			cmd = exec.Command(tarcmd, "-czf", tempdir+"/autowebbackup.tar.gz", s)
		} else {
                        cmd = exec.Command(tarcmd, "-cz", "--newer", fstr, "-f", tempdir+"/autowebbackup.tar.gz", s)
		}
		log.Printf(s)
		err := cmd.Run()
		if err != nil {
                        log.Println(cmd.Path,cmd.Args)
			log.Printf("Tarcmd finished with error: %v", err)
		}
                if do_encrypt {
                        encrypt()
                }

                sparts := strings.SplitAfter(s, "/")
                spart := sparts[len(sparts)-1]

		conn, err := ssh.Dial("tcp", ftpshost+":22", config)
		if err != nil {
        		fmt.Println("Failed to dial: ", err)
        		log.Fatal("Failed to dial: ", err)
		}

		// open an SFTP session over an existing ssh connection.
		client, err := sftp.NewClient(conn)
		if err != nil {
        		fmt.Println(err)
        		log.Fatal(err)
		}

		if daynum == 1 { 
				fmt.Println("Open:", tempdir+transferfile)
				bigFile, err := os.Open(tempdir+transferfile)
				if err != nil {
                        		fmt.Printf("Open file error: %v", err)
				}
				inFile, err := client.Create(monthlydir+"/"+spart+"-"+tstr+"."+transfersuffix)
				if err != nil {
        	        	        fmt.Printf("SFTP create error: %v", err)
				}

				_, err = io.CopyBuffer(inFile, bigFile, buffer)
				if err != nil {
					fmt.Errorf("failed to copy file: %w", err)
				}

				inFile.Close()
				bigFile.Close()
                                time.Sleep(10 * time.Second)
		} else if wd.String() == "Sunday" {
                                fmt.Println("Open:", tempdir+transferfile)
                                bigFile, err := os.Open(tempdir+transferfile)
                                if err != nil {
                                        fmt.Printf("Open file error: %v", err)
                                }
                                inFile, err := client.Create(weeklydir+"/"+spart+"-"+tstr+"."+transfersuffix)
                                if err != nil {
                                        fmt.Printf("SFTP create error: %v", err)
                                }

                                _, err = io.CopyBuffer(inFile, bigFile, buffer)
                                if err != nil {
                                        fmt.Errorf("failed to copy file: %w", err)
                                }

                                inFile.Close()
                                bigFile.Close()
                                time.Sleep(10 * time.Second)
                } else {
                                fmt.Println("Open:", tempdir+transferfile)
                                bigFile, err := os.Open(tempdir+transferfile)
                                if err != nil {
                                        fmt.Printf("Open file error: %v", err)
                                }
                                inFile, err := client.Create(dailydir+"/"+spart+"-"+tstr+"."+transfersuffix)
                                if err != nil {
                                        fmt.Printf("SFTP create error: %v", err)
                                }

                                _, err = io.CopyBuffer(inFile, bigFile, buffer)
                                if err != nil {
                                        fmt.Errorf("failed to copy file: %w", err)
                                }

                                inFile.Close()
                                bigFile.Close()
                                time.Sleep(10 * time.Second)
                }

		os.Remove(tempdir+transferfile)
		client.Close()
		conn.Close()
	}
}

func read_config() {
        err := viper.ReadInConfig() // Find and read the config file
        if err != nil { // Handle errors reading the config file
                log.Fatalf("Config file not found: %v", err)
        }

        ownlog = viper.GetString("own_log")
        if ownlog =="" { // Handle errors reading the config file
                log.Fatalf("Filename for ownlog unknown: %v", err)
        }
// Open log file
        ownlogger = &lumberjack.Logger{
                Filename:   ownlog,
                MaxSize:    5, // megabytes
                MaxBackups: 3,
                MaxAge:     28, //days
                Compress:   true, // disabled by default
        }
//        defer ownlogger.Close()
        log.SetOutput(ownlogger)

        dirs = viper.GetStringSlice("dirs")

        do_trace = viper.GetBool("do_trace")

	tarcmd = viper.GetString("tarcmd")

	ftpsuser = viper.GetString("ftpsuser")
        ftpspassword = viper.GetString("ftpspassword")
        ftpshost = viper.GetString("ftpshost")

        dailydir = viper.GetString("dailydir")
        weeklydir = viper.GetString("weeklydir")
        monthlydir = viper.GetString("monthlydir")
        tempdir = viper.GetString("tempdir")

        dailykeep = viper.GetInt64("dailykeep")
        weeklykeep = viper.GetInt64("weeklykeep")
        monthlykeep = viper.GetInt64("monthlykeep")

	do_encrypt = viper.GetBool("do_encrypt")
	encryptsuffix = viper.GetString("encryptsuffix")
	encryptpassw = viper.GetString("encryptpassw")

	if do_trace {
		log.Println("do_trace: ",do_trace)
		log.Println("own_log; ",ownlog)
		for i, v := range dirs {
			log.Printf("Index: %d, Value: %v\n", i, v )
		}
	}
}

func encrypt() {
    fileSrc, err := os.Open(tempdir+"/autowebbackup.tar.gz")
    if err != nil {
        panic(err)
    }
    defer fileSrc.Close()
    fileDst, err := os.Create(tempdir+"/autowebbackup."+encryptsuffix)
    if err != nil {
        panic(err)
    }
    defer fileDst.Close()
    aes, err := NewAes(32, encryptpassw[0:31])
    if err != nil {
        panic(err)
    }
    err = aes.EncryptStream(fileSrc, fileDst)
    if err != nil {
        panic(err)
    }
    os.Remove(tempdir+"/autowebbackup.tar.gz")
    log.Println("File successfully encrypted")
}

func decrypt() {
    fileSrc, err := os.Open(tempdir+"/autowebbackup."+encryptsuffix)
    if err != nil {
        panic(err)
    }
    defer fileSrc.Close()
    fileDst, err := os.Create(tempdir+"/autowebbackup.tar.gz")
    if err != nil {
        panic(err)
    }
    defer fileDst.Close()
    aes, err := NewAes(32, encryptpassw[0:31])
    if err != nil {
        panic(err)
    }
    err = aes.DecryptStream(fileSrc, fileDst)
    if err != nil {
        panic(err)
    }
    fmt.Println("File successfully decrypted")
}

func myUsage() {
     fmt.Printf("Usage: %s argument\n", os.Args[0])
     fmt.Println("Arguments:")
     fmt.Println("backup        Backup the directories mentioned in the config file")
     fmt.Println("list          List all backups")
     fmt.Println("fetch         Fetch backup from server")
     fmt.Println("decrypt       Decrypt backup")
}

func list() {
config := &ssh.ClientConfig{
        User: ftpsuser,
        Auth: []ssh.AuthMethod{
                ssh.Password(ftpspassword),
        },
        HostKeyCallback: ssh.InsecureIgnoreHostKey(),
}
conn, err := ssh.Dial("tcp", ftpshost+":22", config)
if err != nil {
        fmt.Println("Failed to dial: ", err)
        log.Fatal("Failed to dial: ", err)
}

// open an SFTP session over an existing ssh connection.
client, err := sftp.NewClient(conn)
if err != nil {
        fmt.Println(err)
        log.Fatal(err)
}

// walk a directory
w := client.Walk(dailydir)
for w.Step() {
        if w.Err() != nil {
                continue
        }
        fmt.Println(w.Path())
}

w = client.Walk(weeklydir)
for w.Step() {
        if w.Err() != nil {
                continue
        }
        fmt.Println(w.Path())
}

w = client.Walk(monthlydir)
for w.Step() {
        if w.Err() != nil {
                continue
        }
        fmt.Println(w.Path())
}
client.Close()
conn.Close()
}

func fetch(filename string) {
buffer := make([]byte, 8192)

config := &ssh.ClientConfig{
        User: ftpsuser,
        Auth: []ssh.AuthMethod{
                ssh.Password(ftpspassword),
        },
        HostKeyCallback: ssh.InsecureIgnoreHostKey(),
}
conn, err := ssh.Dial("tcp", ftpshost+":22", config)
if err != nil {
        fmt.Println("Failed to dial: ", err)
        log.Fatal("Failed to dial: ", err)
}

// open an SFTP session over an existing ssh connection.
client, err := sftp.NewClient(conn)
if err != nil {
        fmt.Println(err)
        log.Fatal(err)
}

bigFile, err := os.Create(tempdir+"/autowebbackup."+encryptsuffix)
if err != nil {
    panic(err)
}

inFile, err := client.Open(filename)
if err != nil {
	fmt.Printf("SFTP open error: %v", err)
}

_, err = io.CopyBuffer(bigFile, inFile, buffer)
if err != nil {
	fmt.Errorf("failed to copy file: %w", err)
}

inFile.Close()
bigFile.Close()
client.Close()
conn.Close()
}
