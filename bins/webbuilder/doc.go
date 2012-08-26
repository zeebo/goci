//webbuilder is a program that builds go tests
/*
webbuilder gets all of its arguments from the environment. Here is a summary of
the environment variables it looks for:

	* GOOS: The GOOS to build test binaries for. If unspecified uses the machine's GOOS.
	* GOARCH: The GOARCH to build test binaries for. If unspecified uses the machine's GOARCH.
	* TRACKER: The URL for the tracker. If unspecified uses http://goci.me/rpc/tracker
	* HOSTED: The URL to reach the builder at for sending work. Panics if unspecified.
	* PORT: The port the builder should bind to. Default 9080.

webbuilder does not try to install any tools so you must have everything available
in your path for building go code. This includes git, hg, bzr and go. All binaries
will be stored in temporary directories and removed when done, but some may be left
over in the case of an early exit.
*/
package main
