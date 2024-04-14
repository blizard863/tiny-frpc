package model

type ServiceRunner interface {
	Run() error
	Close() error
}
