package data_test

import (
	"encoding/json"
	"gatso/data"
	"gatso/model"
	"testing"
)

const testDBUri = "mongodb://localhost:27017"
const testOwnerId = 123

var testTaskId string

//initTest will drop the database and create a single test task, belonging to owner 123.
func initTest() *data.MongoDataStore {

	// Use an alternative collection name to prevent other tests running in parrallel spoiling test data.
	ms, err := data.NewMongoDataStore(testDBUri + "#storetest")
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

	return ms
}

func TestNewMongoDataStore(t *testing.T) {
	ms, err := data.NewMongoDataStore(testDBUri)
	defer ms.Close()

	if nil != err {
		t.Error(err)
		return
	}
	if nil == ms {
		t.Errorf("Expected non nil datastore from TestNewMongoDataStore")
		return
	}

	ms, err = data.NewMongoDataStore("")
	if nil == err {
		t.Errorf("Expected exception requesting new datastore with empty uri")
		return
	}
}

func TestMongoDataStore_Close(t *testing.T) {
	ms := initTest()

	c := ms.CountTasks(testOwnerId)
	if c < 0 {
		t.Errorf("expected positive result from count, found %d", c)
		return
	}

	ms.Close()
	// Should return negative result on error state, store being closed.
	c = ms.CountTasks(testOwnerId)
	if c >= 0 {
		t.Errorf("expected negative result from count on closed store, found %d", c)
		return
	}
}

func TestMongoDataStore_AddTask(t *testing.T) {
	ms := initTest()
	defer ms.Close()

	task, err := createTestTask([]byte(`{ "owner": 123, "title": "Another Test Task" }`))
	if nil != err {
		t.Error(err)
		return
	}

	id, err := ms.AddTask(testOwnerId, *task)
	if nil != err {
		t.Error(err)
		return
	}
	if !ms.Exists(id) {
		t.Errorf("Expected new test task with id %s to exist, it does not.", id)
		return
	}
}

func TestMongoDataStore_DeleteTask(t *testing.T) {
	ms := initTest()
	defer ms.Close()

	if !ms.Exists(testTaskId) {
		t.Errorf("Expected test task with id %s to exist, it does not.", testTaskId)
		return
	}

	if !ms.DeleteTask(testOwnerId, testTaskId) {
		t.Errorf("Expected positive result from delete of task %s, found false", testTaskId)
		return
	}

	if ms.Exists(testTaskId) {
		t.Errorf("Expected false from Exists check on id %s, after delete, found true.", testTaskId)
		return
	}
}

func TestMongoDataStore_GetTask(t *testing.T) {
	ms := initTest()
	defer ms.Close()

	task := ms.GetTask(testTaskId)
	if nil == task {
		t.Errorf("Expected result from getTask with task id %s, nil found", testTaskId)
		return
	}

	if task.Id() != testTaskId {
		t.Errorf("Expected result from getTask to match task id %s, found %s", testTaskId, task.Id())
		return
	}

	// Try non existing id
	task = ms.GetTask("madeupid")
	if nil != task {
		t.Errorf("Expected result from getTask with invalid task id to be nil, found %s", task.Id())
		return
	}

	task = ms.GetTask("")
	if nil != task {
		t.Errorf("Expected result from getTask with empty task id to be nil, found %s", task.Id())
		return
	}

}

func TestMongoDataStore_GetTasks(t *testing.T) {
	ms := initTest()
	defer ms.Close()

	tasks, err := ms.GetTasks(testOwnerId)
	if nil != err {
		t.Error(err)
		return
	}
	if len(tasks) != 1 {
		t.Errorf("Expecting result on GetTasks for test owner %d to be 1, found %d", testOwnerId, len(tasks))
		return
	}

	if tasks[0].Id() != testTaskId {
		t.Errorf("Expecting result on GetTasks to contain task %s, found task %s", testTaskId,
			tasks[0].Id())
		return
	}

	// Add a second task and check it returns
	newTask, err := createTestTask([]byte(`{"owner": 123, "title": "yet another new task"}`))
	if nil != err {
		t.Error(err)
		return
	}
	_, err = ms.AddTask(testOwnerId, *newTask)
	if nil != err {
		t.Error(err)
		return
	}

	tasks, err = ms.GetTasks(testOwnerId)
	if nil != err {
		t.Error(err)
		return
	}
	if len(tasks) != 2 {
		t.Errorf("Expecting result on GetTasks for test owner %d to be 2, found %d", testOwnerId, len(tasks))
		return
	}

	// Add a task for another owner and ensure it is NOT included in result
	newTask, err = createTestTask([]byte(`{"owner": 666, "title": "Someone elses business"}`))
	if nil != err {
		t.Error(err)
		return
	}
	_, err = ms.AddTask(666, *newTask)
	if nil != err {
		t.Error(err)
		return
	}

	tasks, err = ms.GetTasks(testOwnerId)
	if nil != err {
		t.Error(err)
		return
	}
	if len(tasks) != 2 {
		t.Errorf("Expecting result on GetTasks for test owner %d to be 2, found %d", testOwnerId, len(tasks))
		return
	}
}

func TestMongoDataStore_UpdateTask(t *testing.T) {
	ms := initTest()
	defer ms.Close()

	task := ms.GetTask(testTaskId)
	if nil == task {
		t.Errorf("Test task %s was not found", testTaskId)
		return
	}

	testNote := "A test note to note is its noted"
	task.Notes = append(task.Notes, testNote)

	err := ms.UpdateTask(testOwnerId, *task)
	if nil != err {
		t.Error(err)
		return
	}

	task = ms.GetTask(testTaskId)
	if nil == task {
		t.Errorf("Test task %s was not found", testTaskId)
		return
	}
	if len(task.Notes) != 1 {
		t.Errorf("Expected one note to be attached to task %s, found %d", task.Id(), len(task.Notes))
		return
	}
	if task.Notes[0] != testNote {
		t.Errorf("Expected test note to be %s, found %s", testNote, task.Notes[0])
		return
	}

}

func TestMongoDataStore_FindTasks(t *testing.T) {
	ms := initTest()
	defer ms.Close()

	query := model.Task{
		Title: "Test Task",
	}
	tasks, err := ms.FindTasks(testOwnerId, query)
	if nil != err {
		t.Error(err)
		return
	}
	if len(tasks) != 1 {
		t.Errorf("Expected one result, found %d", len(tasks))
		return
	}

	// Add another task
	newTask, err := createTestTask([]byte(`{"owner": 123, "title": "Second task"}`))
	if nil != err {
		t.Error(err)
		return
	}
	_, err = ms.AddTask(testOwnerId, *newTask)
	if nil != err {
		t.Error(err)
		return
	}

	// Search for second item
	query = model.Task{Title: "Second task"}
	tasks, err = ms.FindTasks(testOwnerId, query)
	if nil != err {
		t.Error(err)
		return
	}
	if len(tasks) != 1 {
		t.Errorf("Expected one result, found %d", len(tasks))
		return
	}

	// Search for all items
	query = model.Task{Owner: testOwnerId}
	tasks, err = ms.FindTasks(testOwnerId, query)
	if nil != err {
		t.Error(err)
		return
	}
	if len(tasks) != 2 {
		t.Errorf("Expected two results, found %d", len(tasks))
		return
	}

	// Search for non existing
	query = model.Task{Title: "doesn't exist"}
	tasks, err = ms.FindTasks(testOwnerId, query)
	if nil != err {
		t.Error(err)
		return
	}
	if len(tasks) != 0 {
		t.Errorf("Expected no results, found %d", len(tasks))
		return
	}
}

func TestMongoDataStore_FindTasksArrays(t *testing.T) {
	ms := initTest()
	defer ms.Close()

	// Add another task
	newTask, err := createTestTask([]byte(`{"owner": 123, "title": "Second task"}`))
	if nil != err {
		t.Error(err)
		return
	}
	newTask.Labels = append(newTask.Labels, "myLabel")
	_, err = ms.AddTask(testOwnerId, *newTask)
	if nil != err {
		t.Error(err)
		return
	}

	query := model.Task{
		Labels: []string{"myLabel"},
	}

	tasks, err := ms.FindTasks(testOwnerId, query)
	if nil != err {
		t.Error(err)
		return
	}
	if len(tasks) != 1 {
		t.Errorf("Expected one result, found %d", len(tasks))
		return
	}
	if len(tasks[0].Labels) != 1 || tasks[0].Labels[0] != "myLabel" {
		t.Errorf("UnExpected label, found %d", len(tasks))
		return
	}
}

func TestMongoDataStore_GetOthersTasks(t *testing.T) {
	ms := initTest()
	defer ms.Close()

	// Add a task for another owner
	newTask, err := createTestTask([]byte(`{"owner": 666, "title": "Someone elses business"}`))
	if nil != err {
		t.Error(err)
		return
	}
	otherId, err := ms.AddTask(666, *newTask)
	if nil != err {
		t.Error(err)
		return
	}

	tasks, err := ms.GetOthersTasks(testOwnerId)
	if nil != err {
		t.Error(err)
		return
	}
	if len(tasks) != 0 {
		t.Errorf("Expecting empty list of others tasks, found %d items", len(tasks))
		return
	}

	// Not add testowner to others task and see if it appear
	newTask = ms.GetTask(otherId)
	if nil == newTask {
		t.Errorf("Expected other persons task %s, found nothing", otherId)
		return
	}

	newTask.Readers = append(newTask.Readers, testOwnerId)
	if err := ms.UpdateTask(666, *newTask); nil != err {
		t.Error(err)
		return
	}

	tasks, err = ms.GetOthersTasks(testOwnerId)
	if nil != err {
		t.Error(err)
		return
	}
	if len(tasks) != 1 {
		t.Errorf("Expecting others tasks to contain 1 task, found %d items", len(tasks))
		return
	}
	if tasks[0].Id() != otherId {
		t.Errorf("Expecting others tasks to be %s, found %s ", otherId, tasks[0].Id())
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
