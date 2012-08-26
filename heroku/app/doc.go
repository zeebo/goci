//app is the heroku webapp for goci
/*
Environment variables:

	* DIRECTRUN: Set to have goci run tests locally. If set, APP_NAME and API_KEY aren't required but RUNPATH
	* APP_NAME: Name of the app for the runner to send requests. Panics if required and empty.
	* API_KEY: Heroku api key for the runner to send requests. Panics if required and empty.
	* RUNPATH: Path to the github.com/zeebo/goci/runner binary for directrun. Panics if required and empty.
	* DOMAIN: The domain of the hosted page to build absolute urls. Panics if empty.
	* PORT: The port for the webserver to listen on. Panics if empty.
	* DATABASE: URL to the mongo database. Default "mongodb://localhost/gocitest"
	* TEMPLATES: Path to where the templates for the frontend live. Default "./templates"
	* STATIC: Path to where the static files for the frontend live Default "./static"
	* DEBUG: If set, will recompile the templates every invocation.

Accepts the -env argument which if specified will clear the provided environment
and load the environment from the file. Example env file:

	APP_NAME=goci
	API_KEY=foo
	PORT=9080

Be sure to trail the environment file with an empty newline, so for the example
four lines are in the file with the last one == "".
*/
package main
