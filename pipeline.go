package main

type Pipe interface {
	Execute(*Config) error
}
