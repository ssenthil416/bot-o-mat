package input

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

var (
	tk = `[
        {
                "Description": "do the dishes",
                "ETA": 1000
        },
        {
                "Description": "sweep the house",
                "ETA": 3000
        },
        {
                "Description": "do the laundry",
                "ETA": 10000
        },
        {
                "Description": "take out the recycling",
                "ETA": 4000
        },
        {
                "Description": "make a sammich",
                "ETA": 7000
        },
        {
                "Description": "mow the lawn",
                "ETA": 20000
        },
        {
                "Description": "rake the leaves",
                "ETA": 18000
        },
        {
                "Description": "give the dog a bath",
                "ETA": 14500
        },
        {
                "Description": "bake some cookies",
                "ETA": 8000
        },
        {
                "Description": "wash the car",
                "ETA": 20000
        }
	]`

	taskTypes = [6]string{"Unipedal", "Bipedal", "Quadrupedal", "Arachnid", "Radial", "Aeronautical"}
)

//Task Meta Data
type Task struct {
	Description string `json:"Description"`
	Eta         int    `json:"ETA"`
}

//InParams ... Populate input params
type InParams struct {
	NumberOfRobot int      `yaml:"NumOfRobot"`
	RobotData     []string `yaml:"RobotData"`
	UserTask      []string `yaml:"UserTask"`
}

//GetTasksData ... Return default task meta daat
func GetTasksData(tasks *[]Task) (err error) {
	if err := json.Unmarshal([]byte(tk), tasks); err != nil {
		return err
	}
	return nil
}

//ReadYamlFile ... Read User input file
func (inPars *InParams) ReadYamlFile(fn string) (err error) {
	yamlFile, err := ioutil.ReadFile(fn)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(yamlFile, inPars)
	if err != nil {
		return err
	}

	return nil
}

//ValidateInput ... Validate User input
func (inPars InParams) ValidateInput() (err error) {

	if inPars.NumberOfRobot <= 0 || inPars.NumberOfRobot != len(inPars.RobotData) {
		return errors.New("Error : invalid number of Robot or number of Robot is not equal to list of Robot")
	}

	f := true
	for _, rd := range inPars.RobotData {
		if strings.Contains(rd, ":") {
			snt := strings.Split(rd, ":")
			if !checkRobotType(snt[1]) {
				f = false
				break
			}

		} else {
			f = false
			break
		}
	}

	if f == false {
		return errors.New("Error: Robot Type is wrong, check your input")
	}

	f = true
	for _, rd := range inPars.UserTask {
		if !strings.Contains(rd, ":") {
			f = false
			break
		}
	}

	if f == false {
		return errors.New("Error: User Task input is wrong")
	}

	return nil
}

func checkRobotType(rt string) bool {

	for _, t := range taskTypes {
		if rt == t {
			return true
		}
	}
	return false
}
