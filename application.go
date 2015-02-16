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

import (
	"errors"
	"fmt"
)

var (
	ErrApplicationExists = errors.New("The application already exists in marathon, you must update")
)

type Applications struct {
	Apps []Application `json:"apps"`
}

type ApplicationWrap struct {
	Application Application `json:"app"`
}

type Application struct {
	ID            string            `json:"id",omitempty`
	Cmd           string            `json:"cmd,omitempty"`
	Args          []string          `json:"args,omitempty"`
	Constraints   [][]string        `json:"constraints,omitempty"`
	Container     *Container        `json:"container,omitempty"`
	CPUs          float32           `json:"cpus,omitempty"`
	Disk          float32           `json:"disk,omitempty"`
	Env           map[string]string `json:"env,omitempty"`
	Executor      string            `json:"executor,omitempty"`
	HealthChecks  []*HealthCheck    `json:"healthChecks,omitempty"`
	Instances     int               `json:"instances,omitemptys"`
	Mem           float32           `json:"mem,omitempty"`
	Tasks         []*Task           `json:"tasks,omitempty"`
	Ports         []int             `json:"ports,omitempty"`
	RequirePorts  bool              `json:"requirePorts,omitempty"`
	BackoffFactor float32           `json:"backoffFactor,omitempty"`
	Dependencies  []string          `json:"dependencies,omitempty`
	TasksRunning  int               `json:"tasksRunning,omitempty"`
	TasksStaged   int               `json:"tasksStaged,omitempty"`
	User          string            `json:"user,omitempty"`
	Uris          []string          `json:"uris,omitempty"`
	Version       string            `json:"version,omitempty"`
}

func (application *Application) Name(id string) *Application {
	application.ID = id
	return application
}

func (application *Application) CPU(cpu float32) *Application {
	application.CPUs = cpu
	return application
}

func (application *Application) Storage(disk float32) *Application {
	application.Disk = disk
	return application
}

func (application *Application) DependsOn(name string) *Application {
	if application.Dependencies == nil {
		application.Dependencies = make([]string, 0)
	}
	application.Dependencies = append(application.Dependencies, name)
	return application

}

func (application *Application) Memory(memory float32) *Application {
	application.Mem = memory
	return application
}

func (application *Application) Count(count int) *Application {
	application.Instances = count
	return application
}

func (application *Application) Arg(argument string) *Application {
	if application.Args == nil {
		application.Args = make([]string, 0)
	}
	application.Args = append(application.Args, argument)
	return application
}

func (application *Application) AddEnv(name, value string) *Application {
	if application.Env == nil {
		application.Env = make(map[string]string, 0)
	}
	application.Env[name] = value
	return application
}

type ApplicationVersions struct {
	Versions []string `json:"versions"`
}

type ApplicationVersion struct {
	Version string `json:"version"`
}

// Retrieve an array of all the applications which are running in marathon
func (client *Client) Applications() (*Applications, error) {
	applications := new(Applications)
	if err := client.ApiGet(MARATHON_API_APPS, "", applications); err != nil {
		return nil, err
	} else {
		return applications, nil
	}
}

// Retrieve an array of the application names currently running in marathon
func (client *Client) ListApplications() ([]string, error) {
	if applications, err := client.Applications(); err != nil {
		return nil, err
	} else {
		list := make([]string, 0)
		for _, application := range applications.Apps {
			list = append(list, application.ID)
		}
		return list, nil
	}
}

// Checks to see if the application version exists in Marathon
// Params:
// 		name: 		the id used to identify the application
//		version: 	the version (normally a timestamp) your looking for
func (client *Client) HasApplicationVersion(name, version string) (bool, error) {
	if versions, err := client.ApplicationVersions(name); err != nil {
		return false, err
	} else {
		if Contains(versions.Versions, version) {
			return true, nil
		}
		return false, nil
	}
}

// A list of versions which has been deployed with marathon for a specfic application
// Params:
//		name:		the id used to identify the application
func (client *Client) ApplicationVersions(name string) (*ApplicationVersions, error) {
	uri := fmt.Sprintf("%s%s/versions", MARATHON_API_APPS, name)
	versions := new(ApplicationVersions)
	if err := client.ApiGet(uri, "", versions); err != nil {
		return nil, err
	}
	return versions, nil
}

// Change / Revert the version of the application
// Params:
// 		name: 		the id used to identify the application
//		version: 	the version (normally a timestamp) you wish to change to
func (client *Client) SetApplicationVersion(name string, version *ApplicationVersion) (*DeploymentID, error) {
	client.Debug("Changing the application: %s to version: %s", name, version)
	uri := fmt.Sprintf("%s%s", MARATHON_API_APPS, name)
	deploymentId := new(DeploymentID)
	if err := client.ApiPut(uri, version, deploymentId); err != nil {
		client.Debug("Failed to change the application to version: %s, error: %s", version.Version, err)
		return nil, err
	}
	return deploymentId, nil
}

// Retrieve the application configuration from marathon
// Params:
// 		name: 		the id used to identify the application
func (client *Client) Application(name string) (*Application, error) {
	application := new(ApplicationWrap)
	if err := client.ApiGet(fmt.Sprintf("%s%s", MARATHON_API_APPS, name), "", application); err != nil {
		return nil, err
	} else {
		return &application.Application, nil
	}
}

// Validates that the application, or more appropriately it's tasks have passed all the health checks.
// If no health checks exist, we simply return true
// Params:
// 		name: 		the id used to identify the application
func (client *Client) ApplicationOK(name string) (bool, error) {
	/* step: check the application even exists */
	if found, err := client.HasApplication(name); err != nil {
		return false, err
	} else if !found {
		return false, ErrDoesNotExist
	}
	/* step: get the application */
	if application, err := client.Application(name); err != nil {
		return false, err
	} else {
		/* step: if the application has not health checks, just return true */
		if application.HealthChecks == nil || len(application.HealthChecks) <= 0 {
			return true, nil
		}
		/* step: does the application have any tasks */
		if application.Tasks == nil || len(application.Tasks) <= 0 {
			return true, nil
		}

		/* step: iterate the application checks and look for false */
		for _, task := range application.Tasks {
			if task.HealthCheckResult != nil {
				for _, check := range task.HealthCheckResult {
					if !check.Alive {
						return false, nil
					}
				}

			}
		}
		return true, nil
	}
}

// Creates a new application in Marathon
// Params:
// 		application: 		the structure holding the application configuration
func (client *Client) CreateApplication(application *Application) error {
	client.Debug("Creating an application: %s", application)
	return client.ApiPost(MARATHON_API_APPS, &application, nil)
}

// Checks to see if the application exists in marathon
// Params:
// 		name: 		the id used to identify the application
func (client *Client) HasApplication(name string) (bool, error) {
	client.Debug("Checking if application: %s exists in marathon", name)
	if name == "" {
		return false, ErrInvalidArgument
	} else {
		if applications, err := client.ListApplications(); err != nil {
			return false, err
		} else {
			for _, id := range applications {
				if name == id {
					client.Debug("The application: %s presently exist in maration", name)
					return true, nil
				}
			}
		}
		return false, nil
	}
}

// Deletes an application from marathon
// Params:
// 		name: 		the id used to identify the application
func (client *Client) DeleteApplication(name string) error {
	/* step: check of the application already exists */
	client.Debug("Deleting the application: %s", name)
	return client.ApiDelete(fmt.Sprintf("%s%s", MARATHON_API_APPS, name), "", nil)
}

// Performs a rolling restart of marathon application (http://mesosphere.github.io/marathon/docs/rest-api.html#post-/v2/apps/%7Bappid%7D/restart)
// Params:
// 		name: 		the id used to identify the application
func (client *Client) RestartApplication(name string, force bool) (*DeploymentID, error) {
	client.Debug("Restarting the application: %s, force: %s", name, force)
	deployment := new(DeploymentID)
	if err := client.ApiGet(fmt.Sprintf("%s%s/restart", MARATHON_API_APPS, name), "", deployment); err != nil {
		return nil, err
	}
	return deployment, nil
}

// Change the number of instance an application is running
// Params:
// 		name: 		the id used to identify the application
// 		instances:	the number of instances you wish to change to
func (client *Client) ScaleApplicationInstances(name string, instances int) error {
	client.Debug("ScaleApplication: application: %s, instance: %d", name, instances)
	changes := new(Application)
	changes.ID = name
	changes.Instances = instances
	uri := fmt.Sprintf("%s%s", MARATHON_API_APPS, name)
	return client.ApiPut(uri, &changes, nil)
}
