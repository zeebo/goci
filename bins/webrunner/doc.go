//webrunner is a program that runs go tests on the heroku dyno mesh or locally
/*
webrunner gets all of its arguments from the environment. Here is a summary of
the environment variables it looks for (all panics will only occur if the variable is needed):

	* APP_NAME: The name of the heroku app that will be running the tests. Panics if unspecified.
	* API_KEY: The api key of the heroku app that will be runnin the tests. Panics if unspecified.
	* TRACKER: The URL for the tracker. If unspecified uses http://goci.me/rpc/tracker
	* HOSTED: The URL to reach the builder at for sending work. Panics if unspecified.
	* PORT: The port the builder should bind to. Default 9080.
	* DIRECT: If set the runner will run tests locally instead of the heroku dyno mesh.
	* RUNNER: The path to the runner binary for direct running. Panics if unspecified.

In order for webrunner to run tests on heroku, the app must have the binary created
by the import path github.com/zeebo/goci/runner installed to bin/runner. This can
be accomplished by using the github.com/zeebo/buildpack buildpack with a .heroku
config file containing

	+github.com/zeebo/goci/runner

If you would like to run tests directly, set the DIRECT environment variable to
anything. You must also specify the path to the binary created by the import path
github.com/zeebo/goci/runner in the RUNNER environment variable.
*/
package main
