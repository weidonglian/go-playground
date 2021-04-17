package md5

import "crypto/md5"

type Md5Sum [md5.Size]byte

type Md5Result struct {
	path string
	sum  Md5Sum
	err  error
}
