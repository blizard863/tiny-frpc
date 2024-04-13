package model

type ServiceRunner interface {
	Start() error
	Close() error
}
