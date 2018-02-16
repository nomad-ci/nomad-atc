package nomadatc

import "github.com/concourse/atc"

type WorkerResourceType struct {
	atc.WorkerResourceType
	Memory int
}

var CurrentResources = []WorkerResourceType{
	{
		WorkerResourceType: atc.WorkerResourceType{"archive", "concourse/archive-resource", "", false},
		Memory:             512,
	},
	{
		WorkerResourceType: atc.WorkerResourceType{"bosh-deployment", "concourse/bosh-deployment-resource", "", false},
		Memory:             512,
	},
	{
		WorkerResourceType: atc.WorkerResourceType{"bosh-io-release", "concourse/bosh-io-release-resource", "", false},
		Memory:             512,
	},
	{
		WorkerResourceType: atc.WorkerResourceType{"bosh-io-stemcell", "concourse/bosh-io-stemcell-resource", "", false},
		Memory:             512,
	},
	{
		WorkerResourceType: atc.WorkerResourceType{"cf", "concourse/cf-resource", "", false},
		Memory:             512,
	},
	{
		WorkerResourceType: atc.WorkerResourceType{"docker-image", "quay.io/hashicorp/docker-image-resource", "", true},
		Memory:             4096,
	},
	{
		WorkerResourceType: atc.WorkerResourceType{"git", "concourse/git-resource", "", false},
		Memory:             512,
	},
	{
		WorkerResourceType: atc.WorkerResourceType{"github-release", "concourse/github-release-resource", "", false},
		Memory:             512,
	},
	{
		WorkerResourceType: atc.WorkerResourceType{"hg", "concourse/hg-resource", "", false},
		Memory:             512,
	},
	{
		WorkerResourceType: atc.WorkerResourceType{"pool", "concourse/pool-resource", "", false},
		Memory:             512,
	},
	{
		WorkerResourceType: atc.WorkerResourceType{"s3", "concourse/s3-resource", "", false},
		Memory:             512,
	},
	{
		WorkerResourceType: atc.WorkerResourceType{"semver", "concourse/semver-resource", "", false},
		Memory:             256,
	},
	{
		WorkerResourceType: atc.WorkerResourceType{"time", "concourse/time-resource", "", false},
		Memory:             256,
	},
	{
		WorkerResourceType: atc.WorkerResourceType{"tracker", "concourse/tracker-resource", "", false},
		Memory:             512,
	},
}

var AllResources map[string]WorkerResourceType

func init() {
	AllResources = make(map[string]WorkerResourceType)

	for _, br := range CurrentResources {
		AllResources[br.Type] = br
	}
}
