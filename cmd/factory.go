package cmd

import (
	"errors"

	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmrelease "github.com/cloudfoundry/bosh-micro-cli/release"
	bmtar "github.com/cloudfoundry/bosh-micro-cli/tar"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
)

type Factory interface {
	CreateCommand(name string) (Cmd, error)
}

type factory struct {
	commands map[string]Cmd
}

func NewFactory(
	config bmconfig.Config,
	configService bmconfig.Service,
	filesystem boshsys.FileSystem,
	ui bmui.UI,
	extractor bmtar.Extractor,
	releaseValidator bmrelease.Validator,
) Factory {
	return &factory{
		commands: map[string]Cmd{
			"deployment": NewDeploymentCmd(ui, config, configService, filesystem),
			"deploy":     NewDeployCmd(ui, config, filesystem, extractor, releaseValidator),
		},
	}
}

func (f *factory) CreateCommand(name string) (Cmd, error) {
	if f.commands[name] == nil {
		return nil, errors.New("Invalid command name")
	}

	return f.commands[name], nil
}
