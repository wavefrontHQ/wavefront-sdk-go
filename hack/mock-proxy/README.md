# mock-proxy

Use the code in this directory to run a mock of the Wavefront proxy, for manual testing of TLS.
The Makefile generates an internal Certificate Authority and signs a localhost certificate using that CA.
Run the server with `make run`. To destroy all the temporary keyfiles, use `make clean`.