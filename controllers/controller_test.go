package controllers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gatso/controllers"
	"gatso/data"
	"gatso/model"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

const testDBUri = "mongodb://localhost:27017"
const testOwnerId = 123

var testTaskId string
var srv *http.Server

//initControllerTest will drop the database and create a single test task, belonging to owner 123.
func initControllerTest() {

	ms, err := data.NewMongoDataStore(testDBUri)
	if nil != err {
		panic(err)
	}
	ms.Drop()

	task, err := createTestTask([]byte(`{ "owner": 123, "title": "Test Task" }`))
	if nil != err {
		panic(err)
	}
	testTaskId, err = ms.AddTask(testOwnerId, *task)
	if nil != err {
		panic(err)
	}

	ctrl := controllers.NewTaskController(ms)
	mux := http.NewServeMux()
	mux.HandleFunc("/test", ctrl.Tasks)
	mux.HandleFunc("/testothers", ctrl.OthersTasks)

	srv = &http.Server{
		Addr:    ":8008",
		Handler: mux,
	}
	go func() {
		if err := srv.ListenAndServe(); nil != err {
			if err != http.ErrServerClosed {
				panic(err)
			}
		}
	}()
}

func endTest() {
	if nil != srv {
		if err := srv.Close(); nil != err {
			panic(err)
		}
	}
}

func TestTaskControllerTasksGet(t *testing.T) {
	initControllerTest()
	defer endTest()

	resp, err := http.Get(fmt.Sprintf("http://localhost:8008/test?owner=%d", testOwnerId))
	if nil != err {
		t.Errorf("Failed to request TaskController")
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Failed to get expected response.  Expected %s, found %s",
			http.StatusText(http.StatusOK), http.StatusText(resp.StatusCode))
	}
	// Ensure response is valid task
	by, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	var tasks []*model.Task
	if err := json.Unmarshal(by, &tasks); nil != err {
		t.Error(err)
		return
	}
	if nil == tasks || len(tasks) != 1 {
		t.Errorf("Expected one task, found %d", len(tasks))
		return
	}

	// Request with no owner
	resp, err = http.Get("http://localhost:8008/test")
	if nil != err {
		t.Errorf("Failed to request TaskController")
		return
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("Failed to get expected response.  Expected %s, found %s",
			http.StatusText(http.StatusUnprocessableEntity), http.StatusText(resp.StatusCode))
	}

	// Request with invalid owner
	resp, err = http.Get("http://localhost:8008/test?owner=456")
	if nil != err {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Failed to get expected response.  Expected %s, found %s",
			http.StatusText(http.StatusNotFound), http.StatusText(resp.StatusCode))
	}
}

func TestTaskControllerTasksPost(t *testing.T) {
	initControllerTest()
	defer endTest()

	task, err := createTestTask([]byte(`{"owner": 123, "title": "A New Task"}`))
	if nil != err {
		t.Error(err)
		return
	}
	by, err := json.Marshal(&task)
	if nil != err {
		t.Error(err)
		return
	}

	resp, err := http.Post(fmt.Sprintf("http://localhost:8008/test?owner=%d", testOwnerId),
		"application/json", bytes.NewBuffer(by))
	if nil != err {
		t.Error(err)
		return
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Failed to get expected response.  Expected %s, found %s",
			http.StatusText(http.StatusOK), http.StatusText(resp.StatusCode))
	}

	// Check if update succeeded
	resp, err = http.Get(fmt.Sprintf("http://localhost:8008/test?owner=%d", testOwnerId))
	if nil != err {
		t.Error(err)
		return
	}
	by, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	var tasks []*model.Task
	if err := json.Unmarshal(by, &tasks); nil != err {
		t.Error(err)
		return
	}
	if nil == tasks || len(tasks) != 2 {
		t.Errorf("Expected two task, found %d", len(tasks))
		return
	}

	if tasks[1].Title != "A New Task" {
		t.Errorf("Updated task expected title of %s, found %s", "changed", tasks[0].Title)
		return
	}

}

func TestTaskControllerTasksPut(t *testing.T) {
	initControllerTest()
	defer endTest()

	task, err := createTestTask([]byte(`{"owner": 123, "_id": "` + testTaskId + `", "title": "changed"}`))
	if nil != err {
		t.Error(err)
		return
	}
	by, err := json.Marshal(&task)
	if nil != err {
		t.Error(err)
		return
	}

	req, err := http.NewRequest(http.MethodPut,
		fmt.Sprintf("http://localhost:8008/test?owner=%d", testOwnerId), bytes.NewBuffer(by))
	if nil != err {
		t.Error(err)
		return
	}
	req.ContentLength = int64(len(by))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	http.DefaultClient.Timeout = time.Second * 5

	resp, err := http.DefaultClient.Do(req)
	if nil == resp || resp.StatusCode != http.StatusOK {
		t.Errorf("Failed to get expected response.  Expected %s, found %s",
			http.StatusText(http.StatusOK), http.StatusText(resp.StatusCode))
		return
	}

	// Check if update succeeded
	resp, err = http.Get(fmt.Sprintf("http://localhost:8008/test?owner=%d", testOwnerId))
	if nil != err {
		t.Error(err)
		return
	}
	by, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	var tasks []*model.Task
	if err := json.Unmarshal(by, &tasks); nil != err {
		t.Error(err)
		return
	}
	if nil == tasks || len(tasks) != 1 {
		t.Errorf("Expected one task, found %d", len(tasks))
		return
	}
	if tasks[0].Title != "changed" {
		t.Errorf("Updated task expected title of %s, found %s", "changed", tasks[0].Title)
		return
	}
}

func TestTaskControllerTasksDelete(t *testing.T) {
	initControllerTest()
	defer endTest()

	// Add a new task to the owners list
	task, err := createTestTask([]byte(`{"owner": 123, "title": "another test"}`))
	if nil != err {
		t.Error(err)
		return
	}
	by, err := json.Marshal(task)
	if nil != err {
		t.Error(err)
		return
	}
	resp, err := http.Post("http://localhost:8008/test?owner=123", "application/json",
		bytes.NewReader(by))
	if nil != err {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected response %s, found %s",
			http.StatusText(http.StatusCreated), http.StatusText(resp.StatusCode))
		return
	}
	by, err = ioutil.ReadAll(resp.Body)
	if nil != err {
		t.Error(err)
		return
	}
	newId := string(by)

	resp, err = http.Get("http://localhost:8008/test?owner=123")
	if nil != err {
		t.Error(err)
		return
	}
	by, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	var tasks []*model.Task
	err = json.Unmarshal(by, &tasks)
	if nil != err {
		t.Error(err)
		return
	}
	if len(tasks) != 2 {
		t.Errorf("Expected owner to have two tasks, found %d", len(tasks))
		return
	}

	// Now delete the new task
	req, err := http.NewRequest(http.MethodDelete,
		fmt.Sprintf("http://localhost:8008/test?owner=%d&taskId=%s", testOwnerId, newId), nil)
	if nil != err {
		t.Error(err)
		return
	}

	http.DefaultClient.Timeout = time.Second * 5
	resp, err = http.DefaultClient.Do(req)
	if nil == resp || resp.StatusCode != http.StatusOK {
		t.Errorf("Failed to get expected response.  Expected %s, found %s",
			http.StatusText(http.StatusOK), http.StatusText(resp.StatusCode))
		return
	}

	// Check if delete succeeded
	resp, err = http.Get("http://localhost:8008/test?owner=123")
	if nil != err {
		t.Error(err)
		return
	}
	by, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected response %s, found %s",
			http.StatusText(http.StatusOK), http.StatusText(resp.StatusCode))
		return
	}
	err = json.Unmarshal(by, &tasks)
	if nil != err {
		t.Error(err)
		return
	}

	if len(tasks) != 1 {
		t.Errorf("Delete failed")
		return
	}
}

func createTestTask(by []byte) (*model.Task, error) {
	var task model.Task
	if err := json.Unmarshal(by, &task); nil != err {
		return nil, err
	}
	return &task, nil
}
