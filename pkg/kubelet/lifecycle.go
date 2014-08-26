/*
Copyright 2014 Google Inc. All rights reserved.

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

package kubelet

import (
	"fmt"
	"time"
)

// LifecycleOutput is the output from a lifecycle event
type LifecycleOutput struct {
	Details  string
	TimedOut bool
}

// Lifecycle is an interface implemented by things that handle lifecycle events
type Lifecycle interface {
	// PostStart is called after a container is started.  The management system blocks until it returns
	PostStart(containerID string) (LifecycleOutput, error)
	// PreStop is called before a container is stopped.  The management system blocks until it returns
	PreStop(containerID string) (LifecycleOutput, error)
}

func newCommandLineLifecycle(r ContainerCommandRunner) Lifecycle {
	return &commandLineLifecycle{
		runner:  r,
		timeout: time.Second * 30,
	}
}

type commandLineLifecycle struct {
	runner  ContainerCommandRunner
	timeout time.Duration
}

func (c *commandLineLifecycle) runLifecycleCommand(containerID string, cmd []string) (LifecycleOutput, error) {
	kill := make(chan bool)
	timer := time.AfterFunc(c.timeout, func() {
		kill <- true
	})
	data, err := c.runner.RunInContainer(containerID, cmd, kill)
	if timer.Stop() {
		return LifecycleOutput{Details: string(data)}, err
	}
	return LifecycleOutput{
		Details:  "Timed Out",
		TimedOut: true,
	}, fmt.Errorf("aborted")
}

func (c *commandLineLifecycle) PostStart(containerID string) (LifecycleOutput, error) {
	return c.runLifecycleCommand(containerID, []string{"kubernetes-on-start.sh"})
}

func (c *commandLineLifecycle) PreStop(containerID string) (LifecycleOutput, error) {
	return c.runLifecycleCommand(containerID, []string{"kubernetes-on-stop.sh"})
}
