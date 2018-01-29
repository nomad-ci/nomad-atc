package nomadatc

import (
	"code.cloudfoundry.org/lager"
	"github.com/concourse/atc/db"
	"github.com/concourse/atc/engine"
	"github.com/concourse/atc/exec"
)

type BuildDelegateFactory struct {
	Driver     *Driver
	Downstream engine.BuildDelegateFactory
}

func (b *BuildDelegateFactory) Delegate(build db.Build) engine.BuildDelegate {
	ds := b.Downstream.Delegate(build)
	return &BuildDelegate{ds, b.Driver, build.ID()}
}

type BuildDelegate struct {
	engine.BuildDelegate
	driver *Driver
	id     int
}

/*
func (b *BuildDelegate) DBActionsBuildEventsDelegate(plan atc.PlanID) exec.ActionsBuildEventsDelegate {
	return b.downstream.DBActionsBuildEventsDelegate(plan)
}

func (b *BuildDelegate) DBTaskBuildEventsDelegate(plan atc.PlanID) exec.TaskBuildEventsDelegate {
	return b.downstream.DBTaskBuildEventsDelegate(plan)
}

func (b *BuildDelegate) BuildStepDelegate(plan atc.PlanID) exec.BuildStepDelegate {
	return b.downstream.BuildStepDelegate(plan)
}
*/

func (b *BuildDelegate) Finish(logger lager.Logger, err error, success exec.Success, aborted bool) {
	b.driver.CancelForBuild(b.id)
	b.BuildDelegate.Finish(logger, err, success, aborted)
}
