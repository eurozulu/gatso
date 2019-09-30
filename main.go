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
	by.WriteString("\t\tGET Gets the tasks from other users todo lists the owner has access to\n")
	by.WriteString("\t\t    Returns json of all tasks not owned by owner, where owner is a Reader of the task\n")

	return by.Bytes()
}