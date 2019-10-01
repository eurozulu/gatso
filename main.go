package main

import (
	"bytes"
	"fmt"
	"gatso/controllers"
	"gatso/data"
	"net/http"
)

const configDBConnection = "database"
const configPort = "port"
const defaultPort = 8008

func main() {
	cf, err := Newconfig()
	if nil != err {
		panic(err)
	}

	store, err := data.NewMongoDataStore(cf.ReadString(configDBConnection, ""))
	if nil != err {
		panic(err)
	}

	listCtrl := controllers.NewTaskController(store)

	http.HandleFunc("/todo", listCtrl.Tasks)
	http.HandleFunc("/todo/others", listCtrl.OthersTasks)
	http.HandleFunc("/todo/find", listCtrl.Find)
	http.HandleFunc("/todo/help", showApi)
	http.HandleFunc("/health", heartBeatHandler)
	http.HandleFunc("/readiness", heartBeatHandler)

	// Helper mapping for testing (Shouldn't be exposed on a production service)
	http.HandleFunc("/todo/users", listCtrl.Users)

	port := cf.ReadInt(configPort, defaultPort)

	fmt.Printf("Starting todolist on localhost, port %d\n", port)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); nil != err {
		panic(err)
	}

	store.Close()
}
func showApi(w http.ResponseWriter, r *http.Request) {
	w.Write(helpText())
	w.WriteHeader(http.StatusOK)
}
func heartBeatHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func helpText() []byte {
	var by bytes.Buffer
	by.WriteString("Todo API helper\n")
	by.WriteString("\t./todo?owner=nn\n")
	by.WriteString("\t\tGET Gets the todo list for the identified ownerid\n")
	by.WriteString("\t\t    Returns json of all tasks for the given user\n")

	by.WriteString("\t\tPOST Creates a new task in the owners todo list\t<body must have json of task to create by>\n")
	by.WriteString("\t\t     Returns the new task id as the body\n")

	by.WriteString("\t\tPUT \"taskid=ssss\" Update a task in the owners todo list\t<body must have json of task properties to update by>\n")
	by.WriteString("\t\tDELETE \"taskid=ssss\" Delete a task in the owners todo list\n")
	by.WriteString("\t\t       statusOK if delete was carried out.\n")

	by.WriteString("\t./todo/others?owner=nn\n")
	by.WriteString("\t\tGET Gets the tasks from other users todo lists the owner has access to\n")
	by.WriteString("\t\t    Returns json of all tasks not owned by owner, where owner is a Reader of the task\n")

	by.WriteString("\t./todo/find?owner=nn\t<body must have json of task properties to search by>\n")
	by.WriteString("\t\tGET Searches the owners tasks for tasks matching the values given in the query task to\n")
	by.WriteString("\t\t    Body should contain a single task json object containing the values to search for\n")
	by.WriteString("\t\t    String value look for exactly match. Created date will return all tasks create on or after that date.\n")
	by.WriteString("\t\t    Expires date will return all tasks create before that date.\n")
	by.WriteString("\t\t    Array value, notes, labels, readers will match tasks will ALL the given elements of the array in the corrisponding array.\n")

	return by.Bytes()
}
