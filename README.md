# go-playground

## How to Build & Install

Please read the [Golang](https://golang.org/doc/code.html) for details how to organize the workspace. 

Go into a directory of the package, i.e. tree

* Run the file with `func main()`

    ```sh
    $ go run xxxx.go
    ```
  
* Build and check if there is any compilation error

    ```sh
    $ go build
    ``` 

* Install into the bin folder of the $GOPATH

  ```sh
  $ go install
  ```
  
* Run the unit test in each package TestXXX in xxx_test.go
  
  ```sh
  $ go test
  ```