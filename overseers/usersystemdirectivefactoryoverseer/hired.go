package usersystemdirectivefactoryoverseer

import (
	"github.com/herb-go/usersystem"
	"github.com/herb-go/worker"
)

var factoryworker func(loader func(v interface{}) error) (usersystem.Directive, error)
var Team = worker.GetWorkerTeam(&factoryworker)

func GetUserSystemDirectiveFactoryByID(id string) func(loader func(v interface{}) error) (usersystem.Directive, error) {
	w := worker.FindWorker(id)
	if w == nil {
		return nil
	}
	c, ok := w.Interface.(*func(loader func(v interface{}) error) (usersystem.Directive, error))
	if ok == false || c == nil {
		return nil
	}
	return *c
}
