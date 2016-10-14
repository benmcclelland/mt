# mt
[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/benmcclelland/mt) [![build](https://img.shields.io/travis/benmcclelland/mt.svg?style=flat)](https://travis-ci.org/benmcclelland/mt)

golang library for interfacing with magnetic tape device mt command (redhat mt-st-1.1)

Example:
```go
// initialize access to a drive
drive := mt.NewDrive("/dev/nst0")

// rewind media
err := drive.Rewind()
if err != nil {
	return err
}

// position to end of data
err = drive.PositionEOD()
if err != nil {
	return err
}

// back up 5 filemarks
err = drive.BackwardFiles(5)
if err != nil {
	return err
}

// go forward 1 filemark
err = drive.ForwardFiles(1)
if err != nil {
	return err
}
```
