package pipelineserver

import (
	"code.cloudfoundry.org/lager"
	"github.com/concourse/atc/api/auth"
	"github.com/concourse/atc/db"
	"github.com/concourse/atc/engine"
)

type Server struct {
	logger          lager.Logger
	teamFactory     db.TeamFactory
	rejector        auth.Rejector
	pipelineFactory db.PipelineFactory
	engine          engine.Engine
}

func NewServer(
	logger lager.Logger,
	teamFactory db.TeamFactory,
	pipelineFactory db.PipelineFactory,
	engine engine.Engine,
) *Server {
	return &Server{
		logger:          logger,
		teamFactory:     teamFactory,
		rejector:        auth.UnauthorizedRejector{},
		pipelineFactory: pipelineFactory,
		engine:          engine,
	}
}
