package nomadatc

import "github.com/concourse/atc"

var CurrentResources = []atc.WorkerResourceType{
	{"archive", "concourse/archive-resource", "", false},
	{"bosh-deployment", "concourse/bosh-deployment-resource", "", false},
	{"bosh-io-release", "concourse/bosh-io-release-resource", "", false},
	{"bosh-io-stemcell", "concourse/bosh-io-stemcell-resource", "", false},
	{"cf", "concourse/cf-resource", "", false},
	{"docker-image", "concourse/docker-image-resource", "", true},
	{"git", "concourse/git-resource", "", false},
	{"github-release", "concourse/github-release-resource", "", false},
	{"hg", "concourse/hg-resource", "", false},
	{"pool", "concourse/pool-resource", "", false},
	{"s3", "concourse/s3-resource", "", false},
	{"semver", "concourse/semver-resource", "", false},
	{"time", "concourse/time-resource", "", false},
	{"tracker", "concourse/tracker-resource", "", false},
}

var AllResources map[string]atc.WorkerResourceType

func init() {
	AllResources = make(map[string]atc.WorkerResourceType)

	for _, br := range CurrentResources {
		AllResources[br.Type] = br
	}
}
