package main

import (
	"fmt"
	"log"
	"bufio"
	"net/http"
	"time"
	"encoding/json"
	"os"
	"io"
	"flag"
	"sync"
	"sync/atomic"

	udev "github.com/jochenvg/go-udev"
	goserial "github.com/huin/goserial"
	"github.com/gorilla/mux"
)

type Response struct{
	State string `json:"state"`
	TimeLeft uint64 `json:"time_left"`
}

type Request struct{
	AddTime float64 `json:"add_time"`
}

var (
	last uint64 =  0
	deviceEnable bool = false
	lastDefender = &sync.Mutex{}
	c chan struct{} = make(chan struct{}, 1)
	needDesc []string = []string{}
	devicesWriter []io.ReadWriteCloser = []io.ReadWriteCloser{}
	turnOn []byte = []byte("relay on 0\n")
	turnOff []byte = []byte("relay off 0\n")
)

func loadDevicesList(){
	if _,err := os.Stat("devices"); os.IsNotExist(err){
		nfile,err:= os.Create("devices"); if err!=nil{
			fmt.Println(err)
			return
		}
		_,err = nfile.WriteString("c1:8\n")
		nfile.Close()

	}
	file,err := os.Open("devices"); if err!=nil{
		fmt.Println("I hadnt opened file")
		log.Fatal(err)
	}

	defer file.Close()

	scanner:=bufio.NewScanner(file)
	for scanner.Scan(){
		needDesc=append(needDesc, scanner.Text())
	}

	if err:= scanner.Err(); err!=nil{
		log.Fatal(err)
	}
}


func init(){
	loadDevicesList();
	go timer()
	c<-(struct{}{})
}

func timer(){
	for{
		select{
		case <-c:
			lastDefender.Lock()
			if atomic.LoadUint64(&last)!=0{
				atomic.AddUint64(&last,^uint64(1-1))
			//	fmt.Println("Timer lost 1 sec.")
			}else if deviceEnable{
				deviceEnable = false
				go sendToDevice(turnOff)
			}
			lastDefender.Unlock()
			time.Sleep(time.Second)
			c<-(struct{}{})
		}
	}
}

func sendToDevice(data []byte) {

	u:=udev.Udev{}

	//range all devices VID/PID from text file
	for _,deviceData := range needDesc{
		//get data
		d:=u.NewDeviceFromDeviceID(deviceData)
		if d!=nil {
			//Open Port to device
			a:=&goserial.Config{Name:d.Syspath(), Baud:9600}
			s, err :=goserial.OpenPort(a);if err!=nil{
				fmt.Println("Can not open port for device: ",deviceData)
				fmt.Println(err)
				continue
			}
			//send command
			s.Write(data)
			s.Close()
		}else{
			fmt.Println("Device ",deviceData," is undefined")
		}
	}
}


func sendErrorApi(w http.ResponseWriter, errorValue string){
	var x = make(map[string]string)
	x["response"] = errorValue
	fin,_:= json.Marshal(x)
	fmt.Fprintf(w,string(fin))
}

func sendAnswer(w http.ResponseWriter){
	state := "waiting"; if atomic.LoadUint64(&last)!=0{
		state="charging"
        }
        res := Response{State:state, TimeLeft: atomic.LoadUint64(&last)}
        fin, err:= json.Marshal(res);if err!=nil{
		sendErrorApi(w,"Error, can't marshal final data")
                return
        }
        fmt.Fprintf(w, string(fin))
        return
}

func getCharger(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type","application/json")
	sendAnswer(w)
}


func addTime(count uint64, wg *sync.WaitGroup){
	defer wg.Done()
	lastDefender.Lock()
	atomic.AddUint64(&last,count)
        deviceEnable = true
        lastDefender.Unlock()
}

func postCharger(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type","application/json")
	var req Request
	var wg sync.WaitGroup
	decoder:=json.NewDecoder(r.Body)
	err:=decoder.Decode(&req);if err==nil{
		wg.Add(1)
		go addTime(uint64(req.AddTime), &wg)
		wg.Wait()
		sendAnswer(w)
		return
	}
	sendErrorApi(w,`Error to decode your request.`+
	` Are you sure,that request is like {"add_time": 5}`)

}

func main() {
	r:=mux.NewRouter()
	r.HandleFunc("/charger0", getCharger).Methods("GET")
	r.HandleFunc("/charger0", postCharger).Methods("POST")
	wordPort:=flag.String("port", "8888", "string")
	flag.Parse()
	http.Handle("/", r)
	fmt.Println("Started on: 127.0.0.1:",*wordPort)
	log.Fatal(http.ListenAndServe(":"+*wordPort, nil))
}
