package controllers

import (
	"encoding/json"
	"fmt"
	"gatso/data"
	"gatso/model"
	"io/ioutil"
	"net/http"
	"strconv"
)

const paramOwnerId = "owner"
const paramTaskId = "taskId"

type TaskController struct {
	data data.Datastore
}

func NewTaskController(data data.Datastore) *TaskController {
	return &TaskController{data: data}
}

func (c TaskController) Tasks(w http.ResponseWriter, r *http.Request) {

	// request requires the ownerId parameter
	ownerId, err := c.getOwnerId(r)
	if nil != err {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return;
	}

	// delegate the request based on its method

	switch r.Method {
	case http.MethodPost:
		c.createTask(ownerId, w, r)

	case http.MethodGet:
		c.getTasks(ownerId, w, r)

	case http.MethodPut:
		c.updateTask(ownerId, w, r)

	case http.MethodDelete:
		c.deleteTask(ownerId, w, r)

	default:
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return;
	}
}

// OthersTasks retrieves all the task the given ownerId does NOT own, but has been granted view access.
func (c TaskController) OthersTasks(w http.ResponseWriter, r *http.Request) {
	// request requires the ownerId parameter
	ownerId, err := c.getOwnerId(r)
	if nil != err {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return;
	}

	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	tasks, err := c.data.GetOthersTasks(ownerId)
	if nil != err {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if nil == tasks {
		http.Error(w, fmt.Sprintf("user %d not known", ownerId), http.StatusNotFound)
		return
	}

	by, err := json.Marshal(tasks)
	if nil != err {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(by)

}

func (c TaskController) Find(w http.ResponseWriter, r *http.Request) {
	ownerId, err := c.getOwnerId(r)
	if nil != err {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return;
	}
	if nil == r.Body {
		http.Error(w, "No query task in body found", http.StatusUnprocessableEntity)
		return;
	}

	by, err := ioutil.ReadAll(r.Body)
	if nil != err {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var query model.Task
	err = json.Unmarshal(by, &query)
	if nil != err {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tasks, err := c.data.FindTasks(ownerId, query)
	if nil != err {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(tasks) == 0 {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	by, err = json.Marshal(tasks)
	if nil != err {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(by)

}

func (c TaskController) Users(w http.ResponseWriter, r *http.Request) {
	users, err := c.data.Users()
	if nil != err {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(users) == 0 {
		http.Error(w, "No users are defined.  Add a new Task to create the user", http.StatusNotFound)
		return
	}

	by, err := json.Marshal(&users)
	if nil != err {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(by)
}

// getTasks retrieves all the tasks belonging to the given ownerId.
// Only tasks owned by the ownerId are returned.
func (c TaskController) getTasks(ownerId int, w http.ResponseWriter, r *http.Request) {

	tasks, err := c.data.GetTasks(ownerId)
	if nil != err {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if nil == tasks {
		http.Error(w, fmt.Sprintf("user %d not known", ownerId), http.StatusNotFound)
		return
	}

	by, err := json.Marshal(tasks)
	if nil != err {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(by)

}

// createTask will insert a new task under the given owners id.
// the request body must contain a json encoded Task to insert.
// The new task MUST have a title, and an expiry time in the future.
// _id and created times specified in the object are ignored and replaced with the new objects values.
func (c TaskController) createTask(ownerId int, w http.ResponseWriter, r *http.Request) {

	by, err := ioutil.ReadAll(r.Body)
	if nil != err {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	}

	var task model.Task
	err = json.Unmarshal(by, &task)
	if nil != err {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	}

	id, err := c.data.AddTask(ownerId, task)
	if nil != err {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(id))
}

// updateTask updates the task with the _id of the task object given in the request body.
// the body MUST contain a json encoded Task object which, if already existing, must belong to the ownerId.
// If the task already exists, it is replaced with the given object.  If it doesn't exist, it is created.
func (c TaskController) updateTask(ownerId int, w http.ResponseWriter, r *http.Request) {

	by, err := ioutil.ReadAll(r.Body)
	if nil != err {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	var task model.Task
	err = json.Unmarshal(by, &task)
	if nil != err {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err = c.data.UpdateTask(ownerId, task)
	if nil != err {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// deleteTask deletes a task, belonging to the given ownerId.
// the request MUST contain a query parameter containing the taskid and that task must be owned by the given owner id.
func (c TaskController) deleteTask(ownerId int, w http.ResponseWriter, r *http.Request) {

	taskId := r.URL.Query().Get(paramTaskId)
	if taskId == "" {
		http.Error(w, fmt.Sprintf("Missing %s parameter", paramTaskId), http.StatusBadRequest)
		return
	}

	if !c.data.DeleteTask(ownerId, taskId) {
		w.WriteHeader(http.StatusNoContent) // Nothing deleted
		return
	}
	w.WriteHeader(http.StatusOK)
}

// getOwnerId attempts to read the owner ID from a header named [paramOwnerId].
// If not found in the header, it looks on the query URL for the same named parameter
func (c TaskController) getOwnerId(r *http.Request) (int, error) {
	if s, hasOwner := r.Header[paramOwnerId]; hasOwner {
		id, err := strconv.Atoi(s[0])
		if nil != err {
			return -1, err
		}
		return id, nil;
	}

	s := r.URL.Query().Get(paramOwnerId)
	id, err := strconv.Atoi(s)
	if nil != err {
		return -1, fmt.Errorf("Failed to read parameter %s as an owner ID", paramOwnerId)
	}
	return id, nil;
}
