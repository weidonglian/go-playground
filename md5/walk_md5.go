package main

import "crypto/md5"

type md5Sum [md5.Size]byte

type md5Result struct {
	path string
	sum  md5Sum
	err  error
}

func main() {

}
