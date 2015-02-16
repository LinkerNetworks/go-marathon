/*
Copyright 2014 Rohith All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package marathon

type Container struct {
	Type    string    `json:"type,omitempty"`
	Docker  *Docker   `json:"docker,omitempty"`
	Volumes []*Volume `json:"volumes,omitempty"`
}

func (container *Container) Volume(host_path, container_path, mode string) *Container {
	if container.Volumes == nil {
		container.Volumes = make([]*Volume, 0)
	}
	container.Volumes = append(container.Volumes, &Volume{
		ContainerPath: container_path,
		HostPath:      host_path,
		Mode:          mode,
	})
	return container
}

func NewDockerContainer() *Container {
	container := new(Container)
	container.Type = "DOCKER"
	container.Docker = &Docker{
		Image:        "",
		Network:      "BRIDGE",
		PortMappings: make([]*PortMapping, 0),
	}
	container.Volumes = make([]*Volume, 0)
	return container
}

type Volume struct {
	ContainerPath string `json:"containerPath,omitempty"`
	HostPath      string `json:"hostPath,omitempty"`
	Mode          string `json:"mode,omitempty"`
}

type Docker struct {
	Image        string         `json:"image,omitempty"`
	Network      string         `json:"network,omitempty"`
	PortMappings []*PortMapping `json:"portMappings,omitempty"`
}

func (docker *Docker) Container(image string) *Docker {
	docker.Image = image
	return docker
}

func (docker *Docker) Bridged() *Docker {
	docker.Network = "BRIDGE"
	return docker
}

func (docker *Docker) Expose(port int) *Docker {
	docker.ExposePort(port, 0, 0, "tcp")
	return docker
}

func (docker *Docker) ExposeUDP(port int) *Docker {
	docker.ExposePort(port, 0, 0, "udp")
	return docker
}

func (docker *Docker) ExposePort(container_port, host_port, service_port int, protocol string) *Docker {
	if docker.PortMappings == nil {
		docker.PortMappings = make([]*PortMapping, 0)
	}
	docker.PortMappings = append(docker.PortMappings, &PortMapping{
		ContainerPort: container_port,
		HostPort:      host_port,
		ServicePort:   service_port,
		Protocol:      protocol})
	return docker
}

type PortMapping struct {
	ContainerPort int    `json:"containerPort,omitempty"`
	HostPort      int    `json:"hostPort"`
	ServicePort   int    `json:"servicePort,omitempty"`
	Protocol      string `json:"protocol"`
}
