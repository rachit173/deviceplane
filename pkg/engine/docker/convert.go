package docker

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/go-connections/nat"

	"github.com/deviceplane/deviceplane/pkg/engine"
	"github.com/deviceplane/deviceplane/pkg/spec"
	"github.com/docker/docker/api/types/container"
)

func convert(s spec.Service) (*container.Config, *container.HostConfig, error) {
	portBindings, err := ports(s.Ports)
	if err != nil {
		return nil, nil, err
	}
	return &container.Config{
			Cmd:        strslice.StrSlice(s.Command),
			Domainname: s.DomainName,
			Entrypoint: strslice.StrSlice(s.Entrypoint),
			Env:        s.Environment,
			Hostname:   s.Hostname,
			Image:      s.Image,
			Labels:     s.Labels,
			StopSignal: s.StopSignal,
			User:       s.User,
			WorkingDir: s.WorkingDir,
		}, &container.HostConfig{
			CapAdd:         strslice.StrSlice(s.CapAdd),
			CapDrop:        strslice.StrSlice(s.CapDrop),
			DNS:            s.DNS,
			DNSOptions:     s.DNSOpts,
			DNSSearch:      s.DNSSearch,
			ExtraHosts:     s.ExtraHosts,
			GroupAdd:       s.GroupAdd,
			IpcMode:        container.IpcMode(s.Ipc),
			NetworkMode:    container.NetworkMode(s.NetworkMode),
			OomScoreAdj:    int(s.OomScoreAdj),
			PidMode:        container.PidMode(s.Pid),
			PortBindings:   *portBindings,
			Privileged:     s.Privileged,
			ReadonlyRootfs: s.ReadOnly,
			Resources: container.Resources{
				CpusetCpus:        s.CPUSet,
				CPUShares:         int64(s.CPUShares),
				CPUQuota:          int64(s.CPUQuota),
				Memory:            int64(s.MemLimit),
				MemoryReservation: int64(s.MemReservation),
				MemorySwap:        int64(s.MemSwapLimit),
				OomKillDisable:    &s.OomKillDisable, // TODO: this might have the wrong default value
			},
			ShmSize:     int64(s.ShmSize),
			SecurityOpt: s.SecurityOpt,
			UTSMode:     container.UTSMode(s.Uts),
		}, nil
}

func ports(ports []string) (*nat.PortMap, error) {
	_, binding, err := nat.ParsePortSpecs(ports)
	if err != nil {
		return nil, err
	}

	portBindings := nat.PortMap{}
	for k, bv := range binding {
		dcbs := make([]nat.PortBinding, len(bv))
		for k, v := range bv {
			dcbs[k] = nat.PortBinding{
				HostIP:   v.HostIP,
				HostPort: v.HostPort,
			}
		}
		portBindings[nat.Port(k)] = dcbs
	}

	return &portBindings, nil
}

func convertToInstance(c types.Container) engine.Instance {
	return engine.Instance{
		ID:     c.ID,
		Labels: c.Labels,
		// TODO
		Running: c.State == "running",
	}
}
