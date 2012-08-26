//app is the heroku webapp for goci
/*
Environment variables:

	* APP_NAME: Name of the app for the runner to send requests
	* API_KEY: Heroku api key for the runner to send requests
	* DOMAIN: The domain of the hosted page to build absolute urls
	* PORT: The port for the webserver to listen on
	* DIRECTRUN: Set to have goci run tests locally (dangerous. for dev)
	* RUNPATH: Path to the github.com/zeebo/goci/runner binary for directrun
	* DATABASE: URL to the mongo database
	* TEMPLATES: Path to where the templates for the frontend live
	* STATIC: Path to where the static files for the frontend live

Accepts the -env argument which if specified will clear the provided environment
and load the environment from the file. Example env file:

	APP_NAME=goci
	API_KEY=foo
	PORT=9080

Be sure to trail the environment file with an empty newline, so for the example
four lines are in the file with the last one == "".
*/
package main
