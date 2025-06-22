package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Solution struct {
	SolutionID       string                 `json:"SolutionID"`
	Description      string                 `json:"Description"`
	Status           string                 `json:"Status"`
	CreatedTime      string                 `json:"CreatedTime"`
	ExportComponents []Component            `json:"ExportComponents"`
	DynamicData      map[string]interface{} `json:"-"` // Catch all dynamic fields like Topics
}

type Component struct {
	CreatedTime   time.Time `json:"CreatedTime"`
	SolutionID    string    `json:"SolutionID"`
	ComponentType string    `json:"ComponentType"`
	ComponentID   string    `json:"ComponentID"`
}
type Roles []struct {
	Role struct {
		AcessLevels []struct {
			AccessLevel int    `json:"AccessLevel"`
			ObjectID    string `json:"ObjectID"`
			Type        string `json:"Type"`
		} `json:"AcessLevels"`
		ChildRoles  interface{} `json:"ChildRoles"`
		Description string      `json:"Description"`
		RoleName    string      `json:"RoleName"`
		SystemRole  bool        `json:"SystemRole"`
	} `json:"Role"`
}

type Lists []struct {
	List struct {
		ListID   string `json:"ListID"`
		ListName string `json:"ListName"`
		Items    struct {
			Items []struct {
				Value string `json:"Value"`
				Label string `json:"Label"`
			} `json:"Items"`
		} `json:"Items"`
	} `json:"List"`
}

type BizObjects []struct {
	Bizobj struct {
		ObjectName  string `json:"ObjectName"`
		Description string `json:"Description"`
		External    bool   `json:"external"`
	} `json:"Bizobj"`
	Script string `json:"Script"`
}

type ExportComponent struct {
	SolutionID    string `json:"SolutionID"`
	ComponentType string `json:"ComponentType"`
	ComponentID   string `json:"ComponentID"`
}
type ExportedSolution struct {
	SolutionID       string    `json:"SolutionID"`
	Description      string    `json:"Description"`
	Status           string    `json:"Status"`
	CreatedTime      time.Time `json:"CreatedTime"`
	ExportedTime     time.Time `json:"ExportedTime"`
	LastUpdated      time.Time `json:"LastUpdated"`
	CreatedUser      string    `json:"CreatedUser"`
	UpdatedUser      string    `json:"UpdatedUser"`
	ExportedUser     string    `json:"ExportedUser"`
	ExportComponents []ExportComponent
}

type Payload struct {
	SolutionID       string    `json:"SolutionID"`
	Description      string    `json:"Description"`
	Status           string    `json:"Status"`
	CreatedTime      time.Time `json:"CreatedTime"`
	ExportedTime     time.Time `json:"ExportedTime"`
	LastUpdated      time.Time `json:"LastUpdated"`
	CreatedUser      string    `json:"CreatedUser"`
	UpdatedUser      string    `json:"UpdatedUser"`
	ExportedUser     string    `json:"ExportedUser"`
	ExportComponents []payloadcomp
}
type payloadcomp struct {
	ComponentType string `json:"ComponentType"`
	ComponentID   string `json:"ComponentID"`
}
type FinalComp struct {
	CreatedTime   time.Time `json:"CreatedTime"`
	SolutionID    string    `json:"SolutionID"`
	ComponentType string    `json:"ComponentType"`
	ComponentID   string    `json:"ComponentID"`
}

func main() {
	var directory string
	comp := ExportedSolution{}
	finallist := []Component{}

	InitializeLogger()

	args := os.Args
	if len(args) > 1 {
		InfoLogger.Println("Runtime arguments:", args)
		help(args[1])
	}

	cdirectory, _ := os.Getwd()
	directory = filepath.Join(cdirectory, "Solutions")

	InfoLogger.Printf("**********************************************************************************************")
	InfoLogger.Printf("Getting files from %s directory", directory)

	objs, err := os.ReadDir(directory)

	CheckError("Error Reading directory:", err)

	InfoLogger.Println("Got all files from the target directory")

	for _, name := range objs {
		if name.IsDir() || !strings.HasSuffix(name.Name(), ".json") {
			continue
		}
		InfoLogger.Printf("Opening %s file...", name.Name())
		jsonFile, err := os.Open(filepath.Join(directory, name.Name()))
		CheckError("Error Opening file:", err)
		defer jsonFile.Close()

		byteValue, err := io.ReadAll(jsonFile)
		CheckError("Error Reading file:", err)

		var dynamicData map[string]interface{}
		json.Unmarshal(byteValue, &dynamicData)

		jsonString, err := json.Marshal(dynamicData["ExportedSolution"])
		CheckError("Error Marshalling data:", err)

		json.Unmarshal(jsonString, &comp)

		for n := range comp.ExportComponents {
			var finalc Component

			finalc.SolutionID = comp.SolutionID
			finalc.CreatedTime = comp.CreatedTime
			finalc.ComponentType = comp.ExportComponents[n].ComponentType
			finalc.ComponentID = comp.ExportComponents[n].ComponentID

			finallist = append(finallist, finalc)
		}
	}
	if len(finallist) <= 0 {
		InfoLogger.Println("No components to show.")
	} else {
		InfoLogger.Println("Monthly component list is now ready!!")
	}
	InfoLogger.Println("Filtering monthly components...")

	revised := filterLatestComponents(finallist)
	InfoLogger.Println("Constructing payload...")

	payl := CreatePayload(revised)
	CheckError("Error creating payload file:", os.WriteFile("payload.json", payl, 0777))
	InfoLogger.Println("Payload constructed.")

	solid, status := CreateSolution(payl)

	if status == 200 {
		ExportSolution(solid)
	}

}

func CreatePayload(revised []Component) []byte {
	pl := Payload{}
	tempobj := payloadcomp{}
	var solid string
	//fmt.Sprintf("%d", time.Now().Year()) + "_" + time.Now().Month().String()

	InfoLogger.Println("Provide a solutionId based on the tracklist!..")
	fmt.Scan(&solid)
	pl.SolutionID = solid
	pl.Description = solid + "_Monthly_revision"
	pl.Status = "Created"
	pl.CreatedTime = time.Now()
	pl.CreatedUser = "surendhar"

	if len(revised) < 1 {
		fmt.Println("Empty component array")
	}

	fmt.Println("length of array:", len(revised))
	for _, comp := range revised {
		tempobj = payloadcomp{
			ComponentType: comp.ComponentType,
			ComponentID:   comp.ComponentID,
		}
		pl.ExportComponents = append(pl.ExportComponents, tempobj)
	}
	payl, _ := json.Marshal(pl)
	return payl
}

// Function to filter array and keep the latest component by created_time
func filterLatestComponents(components []Component) []Component {
	latestComponents := make(map[string]Component)

	for _, component := range components {

		key := component.ComponentType + "_" + component.ComponentID

		if existingComponent, found := latestComponents[key]; !found || component.CreatedTime.After(existingComponent.CreatedTime) {
			latestComponents[key] = component
		}
	}

	result := make([]Component, 0, len(latestComponents))
	for _, component := range latestComponents {
		result = append(result, component)
	}

	return result
}
func CreateSolution(plbyte []byte) (string, int) {

	var pl Payload

	CheckError("Error loading env file:", godotenv.Load())

	json.Unmarshal(plbyte, &pl)

	InfoLogger.Println("Calling API:", os.Getenv("create"))

	req, err := http.NewRequest("POST", os.Getenv("create"), bytes.NewBuffer(plbyte))

	CheckError("Error in creating solution:", err)

	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("X-API-KEY", os.Getenv("X_API_KEY"))

	client := &http.Client{}
	resp, err := client.Do(req)
	CheckError("Error calling api", err)
	defer resp.Body.Close()

	InfoLogger.Println("Status Code:", resp.StatusCode)
	//fmt.Println("Response Body:", string(body))

	return pl.SolutionID, resp.StatusCode

}

func ExportSolution(solid string) {

	var exp_url string

	if strings.HasSuffix(os.Getenv("export"), "/") {
		exp_url = os.Getenv("export") + solid
	} else {
		exp_url = os.Getenv("export") + "/" + solid
	}
	InfoLogger.Println("Calling API:", exp_url)
	req, err := http.NewRequest("POST", exp_url, nil)

	CheckError("Error in exporting solution:", err)

	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("X-API-KEY", os.Getenv("X_API_KEY"))

	client := &http.Client{}
	resp, err := client.Do(req)
	CheckError("Error calling api", err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	CheckError("Error converting Response:", err)

	InfoLogger.Println("Status Code:", resp.StatusCode)
	CheckError("Error creating solution file:", os.WriteFile(solid+".json", body, 0777))
	InfoLogger.Println("Solution file created.")
}
func help(arg string) {

	hlp := `
	
	Application to consolidate e-invois solutions and create only one solution with the updated components
	
	Points To Remember:
	
	
	.env 		- file is mandatory to maintain cral URLs and api key.
	Solutions 	- is the mandatory directory where all solutions have to be placed.`

	errarg := `
	
	Please provide a valid argument.
	
	-h or -?	- for more details.`

	switch arg {
	case "-h", "-?":
		InfoLogger.Fatal(hlp)
	default:
		InfoLogger.Fatal(errarg)
	}

}
