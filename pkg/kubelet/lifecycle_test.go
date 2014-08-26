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
	"testing"
	"time"
)

type commandState struct {
	id      string
	command []string
}

type fakeCommandRunner struct {
	commands []commandState
	sleep    *time.Duration
	output   []byte
	err      error
}

func (f *fakeCommandRunner) RunInContainer(id string, cmd []string, kill <-chan bool) ([]byte, error) {
	if f.sleep != nil {
		select {
		case <-time.After(*f.sleep):
		case <-kill:
			return f.output, fmt.Errorf("aborted")
		}
	}
	f.commands = append(f.commands, commandState{id: id, command: cmd})
	return f.output, f.err
}

func TestPostStart(t *testing.T) {
	fake := fakeCommandRunner{
		output: []byte("Foo"),
	}
	lifecycle := newCommandLineLifecycle(&fake)
	out, err := lifecycle.PostStart("foo")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if out.Details != "Foo" {
		t.Errorf("unexpected output: %v", out)
	}

	if len(fake.commands) != 1 || fake.commands[0].id != "foo" || fake.commands[0].command[0] != "kubernetes-on-start.sh" {
		t.Errorf("unexpected commands: %v", fake.commands)
	}
}

func TestPreStop(t *testing.T) {
	fake := fakeCommandRunner{
		output: []byte("Foo"),
	}
	lifecycle := newCommandLineLifecycle(&fake)
	out, err := lifecycle.PreStop("foo")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if out.Details != "Foo" {
		t.Errorf("unexpected output: %v", out)
	}

	if len(fake.commands) != 1 || fake.commands[0].id != "foo" || fake.commands[0].command[0] != "kubernetes-on-stop.sh" {
		t.Errorf("unexpected commands: %v", fake.commands)
	}
}

func TestTimeout(t *testing.T) {
	sleep := time.Second * 10
	fake := fakeCommandRunner{
		sleep: &sleep,
	}
	lifecycle := newCommandLineLifecycle(&fake)
	lifecycle.(*commandLineLifecycle).timeout = time.Millisecond
	_, err := lifecycle.PreStop("foo")

	if err == nil {
		t.Error("unexpected non-error")
	}
}
