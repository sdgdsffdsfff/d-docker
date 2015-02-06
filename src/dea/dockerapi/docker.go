package dockerapi

import (
	steno "github.com/cloudfoundry/gosteno"
	"dea-docker/src/dea/config"
	"icode.jd.com/cdlxyong/go-dockerclient"
	"fmt"
	"strconv"
	"strings"
	"errors"
	"bytes"
)

const(
	 minShares = 2
	 sharesPerCPU = 1024
	 milliCPUToCPU = 1000
	)

type Docker struct {
	logger				*steno.Logger
	conf				*config.Config
	dockerClient		*docker.Client
}

func NewDocker (f *config.Config) (*Docker, error) {
	dockerClient, err := docker.NewClient(f.Docker.Url)
	
	if err != nil {
		return nil, errors.New(fmt.Sprintf("NewDocker fail,%s", err) )
	}
	return &Docker{
		logger:				steno.NewLogger("dea-docker"),
		conf:				f,
		dockerClient:		dockerClient,
	},nil
}

//stop container
func (d *Docker) StopContainer(containerId string) error {
	
	if containerId == "" {
		return errors.New("StopContainer fail, containerId is empty")	
	}
	
	err := d.dockerClient.StopContainer(containerId, 2)
	
	if err != nil {
		return err	
	}
	
	return nil
}

//remove Image
func (d *Docker) RemoveImage (imageName string, imageTarg string) error {
	
	err := d.dockerClient.RemoveImage(fmt.Sprintf("%s:%s", imageName, imageTarg) )
	if err != nil {
		d.logger.Errorf("Docker Remove image:{%s} fail, %s", fmt.Sprintf("%s:%s", imageName, imageTarg), err.Error() )	
		return err
	}
	
	return nil
}

//remove container
func (d *Docker) RemoveContainer(containerId string) error {
	
	opts := docker.RemoveContainerOptions{ID: containerId,}
	
	err := d.dockerClient.RemoveContainer(opts)
	
	if err != nil {
		d.logger.Errorf("Docker Remove Container:{%s} fail, %s", containerId, err.Error() )	
		return err
	}
	
	return nil
}


func (d *Docker) ContainerInfo(id string) (*docker.Container, error) {
	
	if id == "" {
		return nil, errors.New("Container info fail, containerId is empty")	
	}
	container, err := d.dockerClient.InspectContainer(id)
	
	if err != nil {
		return nil, err	
	}
	
	return container, nil
}

//check image is exists return true/false
func (d *Docker) ExistsImage(image string, targ string) (bool, error) {
	
	if image == "" {
		return false,errors.New("Docker.ExistsImage and image is empty")	
	}
	
	images, err := d.dockerClient.ListImages(true)
	
	if err != nil {
		return false, err	
	}
	
	checkImage := fmt.Sprintf("%s:%s", image, targ)
	
	if images != nil && len(images) >0 {
		
		for _,	image := range images {
			repoTags := image.RepoTags
			if repoTags == nil || len(repoTags) == 0 {
				continue	
			} 
			
			for _, repotage := range repoTags {
				if repotage == checkImage {
					return true, nil	
				}	
			}
			
		}
	}
	
	return false,nil
}

//pull image
// parameter image name, image targ ,image registry
func (d *Docker) PullImage(image string, targ string, registry string) error {
	
	if image == "" {
		return errors.New(fmt.Sprintf("Docker.PullImage{%s} fail, image is empty", image) )	
	}
	if registry == "" {
		d.logger.Warnf(fmt.Sprintf("Docker.PullImage{%s} fail, registry is empty, and use default registry", image))
	}
	
	var opt docker.PullImageOptions
	var buf bytes.Buffer
	
	if targ != "" {
		opt = docker.PullImageOptions{Repository: image,Tag: targ,OutputStream: &buf, Registry: registry,}
	}else {
		opt = docker.PullImageOptions{Repository: image, OutputStream: &buf, Registry: registry,}
	}
	
	err := d.dockerClient.PullImage(opt, docker.AuthConfiguration{})
	
	if err != nil {
		return err	
	}
	
	d.logger.Infof("Docker.PullImage{image:%s,targ:%s,registry:%s} response:%s", image, targ, registry, buf.String() )
	
	return nil
}

//create container 
//run container
func (d *Docker) RunContainer(container *Container, netMode string) (*Container, error) {
	
	envVariables :=	d.makeEnvironmentVariables(container)
	binds := d.makeBinds(container)
	exposedPorts, portBindings := d.makePortsAndBindings(container)
	
	opts := docker.CreateContainerOptions{
		Name: container.Name,
		Config: &docker.Config{
			Cmd:          container.Command,
			Env:          envVariables,
			ExposedPorts: exposedPorts,
			Hostname:     "",
			Image:        container.Image,
			Memory:       int64(container.Memory),
			CPUShares:    int64(d.milliCPUToShares(container.CPU)),
			WorkingDir:   container.WorkingDir,
		},
	}
	
	dockerContainer, err := d.dockerClient.CreateContainer(opts)
	if err != nil {
		d.logger.Errorf("dockerClient CreateContainer fail,%s", err)
		return nil, err
	}
	
	//starting
	
	err = d.dockerClient.StartContainer(dockerContainer.ID, &docker.HostConfig{
		PortBindings: portBindings,
		Binds:        binds,
		NetworkMode:  netMode,
		Privileged:   false,		
	})
	
	if err != nil {
		d.logger.Errorf("dockerClient startContainer fail,%s", err)
		return nil,err	
	}
	container.ID = dockerContainer.ID
	return container, nil
}

func (d *Docker) makeBinds(container *Container) []string {
	binds := []string{}
	for _, mount := range container.VolumeMounts {
		b := fmt.Sprintf("%s:%s", mount.DstPath, mount.MountPath)
		if mount.ReadOnly {
			b += ":ro"
		}
		binds = append(binds, b)
	}
	return binds
}

func (d *Docker)makeEnvironmentVariables(container *Container) []string {
	var result []string
	for _, value := range container.Env {
		result = append(result, fmt.Sprintf("%s=%s", value.Name, value.Value))
	}
	return result
}

func (d *Docker) makePortsAndBindings(container *Container) (map[docker.Port]struct{}, map[docker.Port][]docker.PortBinding) {
	exposedPorts := map[docker.Port]struct{}{}
	portBindings := map[docker.Port][]docker.PortBinding{}
	for _, port := range container.Ports {
		exteriorPort := port.HostPort
		if exteriorPort == 0 {
			// No need to do port binding when HostPort is not specified
			continue
		}
		interiorPort := port.ContainerPort
		// Some of this port stuff is under-documented voodoo.
		// See http://stackoverflow.com/questions/20428302/binding-a-port-to-a-host-interface-using-the-rest-api
		var protocol string
		switch strings.ToUpper(string(port.Protocol)) {
		case "UDP":
			protocol = "/udp"
		case "TCP":
			protocol = "/tcp"
		default:
			d.logger.Warnf("Unknown protocol '%s': defaulting to TCP", port.Protocol)
			protocol = "/tcp"
		}
		dockerPort := docker.Port(strconv.Itoa(interiorPort) + protocol)
		exposedPorts[dockerPort] = struct{}{}
		portBindings[dockerPort] = []docker.PortBinding{
			{
				HostPort: strconv.Itoa(exteriorPort),
				HostIP:   port.HostIP,
			},
		}
	}
	return exposedPorts, portBindings
}

func (d *Docker) milliCPUToShares(milliCPU int) int {
	if milliCPU == 0 {
		// zero milliCPU means unset. Use kernel default.
		return 0
	}
	// Conceptually (milliCPU / milliCPUToCPU) * sharesPerCPU, but factored to improve rounding.
	shares := (milliCPU * sharesPerCPU) / milliCPUToCPU
	if shares < minShares {
		return minShares
	}
	return shares
}