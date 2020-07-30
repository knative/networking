# Flaky test image

The image contains a simple Go webserver, `main.go`, that will only succeed
every Nth request. The value of N is specified in the PERIOD environment
variable.

## Trying out

To run the image as a Service outisde of the test suite:

`ko apply -f service.yaml`
