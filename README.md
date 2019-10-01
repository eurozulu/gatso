Simple todo list

Lists are defined by owners, each owner having a unique id.
Each owner has one list, with one or more tasks in it.

Installation
kubernetes install:
using default node name: docker-desktop to select the correct node to deploy the pod.
Update the mongodb-deployment.yml file, kubernetes.io/hostname to reflect your local nodename.

The mongodb-deployment.yml deploys a headless service and a StatefulSet to provide
mongodb as a service on "host" database, port 27017

todolist-deployment.yml deploys a simple, 1 replica pod containing to todo service.

The service can be exposed to the localhost via a expose service set in the mkexposesvc.sh script.


Local install
To install as a local pair of docker containers, use a standard mongo container
and run the mkdocker.sh script to generate the tolist image. (This attempts to push the image to a public repo)

The local container should expose port 8008 to the docker bridge.

Configuration
Service has two properties to configure:
"database"	The mongodb connection string in the form mongodb://<user>:<password>@database:27017
"port" 		The local port the service will listen on for inbound http requests, default is 8008.

These properties are in the todo-properties.json file, found in the same location as the service executable
(Or in a location specified by the TODOHOME environment variable)


REST Api root url:  http://localhost/todo
(curl http://localhost/todo/help to get a list of available end points)


Security:
The service is not secure in any way.  Non encrypted/TLS endpoints are used to simplify testing.
insecure username/passwords in plain text config and deployment files.
No additional checks are carried out on inbound requests.
No authentication / authorisation enabled on endpoints.

Minimal dependencies
Dependencies have been kept to a minimum to keep the project simple.
Production code would employ third party libaries for some aspects such as:
http mux and json as the regular packages are not very performant.
