package service

type Tasker interface {
	DeleteFile(uuid string) error
	DeleteDir(uuid string) error
}
