package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	in "sample/botomat/input"
)

const (
	maxTaskPerWorker = 5
	healthEndPoint   = "/health"
	statusEndPoint   = "/status"
)

type rstat struct {
	Name   string `json:"Name"`
	Status string `json:"Status"`
}

type stat struct {
	Info  string  `json:"Info"`
	Robot []rstat `json:"Robot"`
}

type wTaskInfo struct {
	Name   string    `json:"name"`
	Tasks  []in.Task `json:"tasks"`
	Status []bool    `json:"status"`
}

var (
	userYamlFile string
	wmd          = make(map[string]wTaskInfo)
	tasks        = make([]in.Task, 0, 1)
	currStat     = stat{}
	maxGoCPU     = 0
	currGoRout   = 0
	mutex        = &sync.Mutex{}
	logF, _      = os.Create("/tmp/botomat.log")
	mlog         = log.New(logF, "", 0)
)

func init() {
	maxGoCPU = runtime.NumCPU()
}

func main() {
	flag.StringVar(&userYamlFile, "userYamlFile", "", "File name : Enter path and file name for user input")
	flag.Parse()

	if userYamlFile == "" {
		mlog.Println("Error : Input files missing ")
		flag.Usage()
		return
	}

	inParams := in.InParams{}

	//Read input file
	err := inParams.ReadYamlFile(userYamlFile)
	if err != nil {
		mlog.Println("Error : In Reading User input : ", err)
		return
	}

	//Validate input params
	if err := inParams.ValidateInput(); err != nil {
		mlog.Println(err)
		return
	}

	mlog.Printf("inParams = %+v\n", inParams)

	//Get default task data
	if err := in.GetTasksData(&tasks); err != nil {
		mlog.Printf("Error in getting tasks data : %v\n", err)
		return
	}

	//Get User Data and merge with default task
	for _, ts := range inParams.UserTask {
		tmp := in.Task{}
		sp := strings.Split(ts, ":")
		tmp.Description = sp[0]
		tmp.Eta, _ = strconv.Atoi(sp[1])
		tasks = append(tasks, tmp)
	}

	//Wait group
	wg := &sync.WaitGroup{}

	//Comm between GoRoutine
	inChan := make(chan wTaskInfo)
	outChan := make(chan wTaskInfo)

	//ProgressDashboard Worker
	go startWebServer()

	//Status or progress checker
	wg.Add(1)
	go statWorker(wg, inParams, outChan)

	//Reduce maxGoCPU, Since we started one go routine here
	maxGoCPU = maxGoCPU - 2

	maxGoProc := 0
	if inParams.NumberOfRobot > maxGoCPU {
		maxGoProc = inParams.NumberOfRobot - maxGoCPU
	} else {
		maxGoProc = inParams.NumberOfRobot
	}
	runtime.GOMAXPROCS(maxGoProc)

	routCount := 0
	currGoRout = 0
	//Start Worker Routine
	for {
		wi := wTaskInfo{}
		getTaskList(tasks, &wi)

		//Create task name and go routine
		wName := fmt.Sprintf("Task-%d", routCount+1)
		wi.Name = wName
		wg.Add(1)
		routCount = routCount + 1
		mutex.Lock()
		wmd[wName] = wi
		currGoRout = currGoRout + 1
		mutex.Unlock()
		go worker(wg, inChan, outChan)
		inChan <- wi

		//Waiting for other worker routinue
		if routCount == inParams.NumberOfRobot {
			break
		} else if currGoRout < maxGoCPU {
			continue
		} else if currGoRout == maxGoCPU && routCount < inParams.NumberOfRobot {

			//Wait to complete other tasks to complete
			for {
				if currGoRout < maxGoCPU {
					break
				}
				mlog.Println("Waiting for other task to complete, to create remaing task!!!!!!")
				time.Sleep(500 * time.Millisecond)
			}
		}
	}

	wg.Wait()

	//Final Report
	mlog.Println("\nFinal Report")
	report()
}

func report() {
	tmp := make(map[string]wTaskInfo)
	mutex.Lock()
	tmp = wmd
	mutex.Unlock()
	cs := stat{}
	cs.Info = fmt.Sprintf("\n\nReport RoBot Status @ %s\n", time.Now().String())
	rs := rstat{}
	for k, v := range tmp {
		lenStat := len(v.Status)
		lenTask := len(v.Tasks)
		if lenTask > 0 {
			lenTask = lenTask - 1
		}

		rs.Name = fmt.Sprintf("%s  ", k)
		if lenStat == maxTaskPerWorker {
			rs.Status = fmt.Sprintf("Completed %d Tasks\n", lenStat)
		} else {
			rs.Status = fmt.Sprintf("Working on Task : %s : Completed %d Tasks\n", v.Tasks[lenTask].Description, lenStat)
		}
		cs.Robot = append(cs.Robot, rs)
	}
	currStat = cs
	mlog.Printf("Report : %+v\n", currStat)
}

func startWebServer() {
	//Http router
	http.HandleFunc(healthEndPoint, GetHealthCheck)
	http.HandleFunc(statusEndPoint, GetStatus)
	if err := http.ListenAndServe(":3222", nil); err != nil {
		fmt.Println("ListenAndServe: ", err)
		return
	}
}

//GetHealthCheck ...
func GetHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

//GetStatus ...
func GetStatus(w http.ResponseWriter, r *http.Request) {
	mlog.Println("inside getStatus")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	mutex.Lock()
	json.NewEncoder(w).Encode(currStat)
	mutex.Unlock()
}

func worker(wg *sync.WaitGroup, inChan <-chan wTaskInfo, outChan chan<- wTaskInfo) {
	defer wg.Done()

	wi := <-inChan
	for _, ts := range wi.Tasks {
		/*
			ticker := time.NewTicker(time.Minute)
			go func() {
				for t := range ticker.C {
					fmt.Printf("Working on task : %s @ ", ts.Description)
					fmt.Println(t)
				}
			}()
		*/
		time.Sleep(time.Duration(ts.Eta) * time.Millisecond)
		//ticker.Stop()
		wi.Status = append(wi.Status, true)
		outChan <- wi
		//mlog.Printf("Robot  %s :: Done with task : %s\n",wi.Name, ts.Description)
	}
	mutex.Lock()
	currGoRout = currGoRout - 1
	mutex.Unlock()
	mlog.Printf("Robot %s : is done\n", wi.Name)
}

func statWorker(wg *sync.WaitGroup, inParams in.InParams, outChan <-chan wTaskInfo) {
	wg.Done()

	ct := 0
	//Every Five Minute display worker status
	ticker := time.NewTicker(500 * time.Millisecond)
	go func() {
		for t := range ticker.C {
			mlog.Println(t.String())
			report()
		}
	}()

	//Update Current status
	for {
		wi := <-outChan

		//Update status
		mutex.Lock()
		wmd[wi.Name] = wi
		mutex.Unlock()

		if len(wi.Status) == maxTaskPerWorker {
			ct = ct + 1
		}

		if ct == inParams.NumberOfRobot {
			ticker.Stop()
			break
		}
	}
}

func getTaskList(tasks []in.Task, wi *wTaskInfo) {
	min := 0
	max := len(tasks) - 1

	wi.Tasks = make([]in.Task, 0, 1)
	wi.Status = make([]bool, 0, 1)
	for i := 0; i < maxTaskPerWorker; i++ {
		t := tasks[random(min, max)]
		wi.Tasks = append(wi.Tasks, t)
	}
	return
}

func random(min, max int) int {
	return rand.Intn(max-min) + min
}
