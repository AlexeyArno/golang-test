# Power Changer

## Install
###### 1. Install `libudev`, `gcc`
###### 2. `go get github.com/AlexeyArno/golang-usb-test`
###### 3. `cd $GOPATH/src/github.com/AlexeyArno/golang-usb-test`
###### 4. `go get ./`
###### 5. `go build`
###### 6. `./golang-usb-test -port=1234`
## Describe

###### A simple program that sends commands: "ready on 0 \ n", "ready off 0 \ n" to the serial ports defined in the text file, after the timeout expires. Add and receive information using POST and GET http requests:
`(POST) 127.0.0.1:1234:charger0 -> ({"add_time": 5}) => {"state": "waiting": "charging", "time_left": 10}`
`(GET) 127.0.0.1:1234:charger0 => {"state": "waiting": "charging", "time_left": 10}`