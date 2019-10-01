<h1>Simple todo list</h1>
<p />
<p>
Lists are defined by owners, each owner having a unique id.
Each owner has one list, with one or more tasks in it.
</p>

<h3>Installation</h3>
<p>kubernetes install:<br />
uses default node name: <code>docker-desktop</code> to select the correct node to deploy the pod.<br/>
Update the <code>mongodb-deployment.yml</code> file, <code>kubernetes.io/hostname</code> to reflect your local/cluster nodename.
</p><p>
The <code>mongodb-deployment.yml</code> deploys a headless service and a StatefulSet to provide
mongodb as a service on hostname <code>database</code>, port 27017
</p>
<p>
<code>todolist-deployment.yml</code> deploys a simple, 1 replica pod containing the todo service.

The service can be exposed to the localhost via an expose service set with the <code>mkexposesvc.sh</code> script.
</p>

<p>Local install:<br />
To install as a local pair of docker containers, use a standard mongo container
and run the mkdocker.sh script to generate the todolist image. (This attempts to push the image to a public repo)

The local container should expose port 8008 to the docker bridge.
</p>

<p>Configuration<br/>
Service has two properties to configure:<br/>
<code>database</code>	The mongodb connection string in the form <code>mongodb://<user>:<password>@database:27017</code><br/>
<code>port</code> 		The local port the service will listen on for inbound http requests, default is 8008.

These properties are in the todo-properties.json file, found in the same location as the service executable
(Or in a location specified by the TODOHOME environment variable)
</p>

<p>
REST Api root url:  http://localhost/todo<br/>
(curl http://localhost/todo/help to get a list of available end points)
</p>
<p>
Security:<br/>
The service is not secure in any way.  Non encrypted/TLS endpoints are used to simplify testing.<br/>
insecure username/passwords in plain text config and deployment files.<br/>
No additional checks are carried out on inbound requests.<br/>
No authentication / authorisation enabled on endpoints.<br/>

</p>
<p>Minimal dependencies<br/>
Dependencies have been kept to a minimum to keep the project simple.
Production code would employ third party libaries for some aspects such as:
http mux and json as the regular packages are not very performant.
</p>