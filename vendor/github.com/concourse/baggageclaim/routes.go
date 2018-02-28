package baggageclaim

import "github.com/tedsuo/rata"

const (
	ListVolumes   = "ListVolumes"
	GetVolume     = "GetVolume"
	CreateVolume  = "CreateVolume"
	DestroyVolume = "DestroyVolume"

	CreateVolumeAsync       = "CreateVolumeAsync"
	CreateVolumeAsyncCancel = "CreateVolumeAsyncCancel"
	CreateVolumeAsyncCheck  = "CreateVolumeAsyncCheck"

	SetProperty   = "SetProperty"
	SetTTL        = "SetTTL"
	SetPrivileged = "SetPrivileged"
	StreamIn      = "StreamIn"
	StreamOut     = "StreamOut"
)

var Routes = rata.Routes{
	{Path: "/volumes", Method: "GET", Name: ListVolumes},
	{Path: "/volumes", Method: "POST", Name: CreateVolume},

	{Path: "/volumes-async", Method: "POST", Name: CreateVolumeAsync},
	{Path: "/volumes-async/:handle", Method: "GET", Name: CreateVolumeAsyncCheck},
	{Path: "/volumes-async/:handle", Method: "DELETE", Name: CreateVolumeAsyncCancel},

	{Path: "/volumes/:handle", Method: "GET", Name: GetVolume},
	{Path: "/volumes/:handle/properties/:property", Method: "PUT", Name: SetProperty},
	{Path: "/volumes/:handle/ttl", Method: "PUT", Name: SetTTL},
	{Path: "/volumes/:handle/privileged", Method: "PUT", Name: SetPrivileged},
	{Path: "/volumes/:handle/stream-in", Method: "PUT", Name: StreamIn},
	{Path: "/volumes/:handle/stream-out", Method: "PUT", Name: StreamOut},
	{Path: "/volumes/:handle", Method: "DELETE", Name: DestroyVolume},
}
