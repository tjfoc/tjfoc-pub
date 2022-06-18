/*
Copyright Suzhou Tongji Fintech Research Institute 2018 All Rights Reserved.

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

package container

import (
	"fmt"
	"io"
	"sync"

	"github.com/tjfoc/tjfoc/core/common/flogging"
	"github.com/tjfoc/tjfoc/core/container/api"
	"github.com/tjfoc/tjfoc/core/container/dockercontroller"
	"golang.org/x/net/context"
)

type refCountedLock struct {
	refCount int
	lock     *sync.RWMutex
}

var vmLogger = flogging.MustGetLogger("controller")

//VMController - manages VMs
//   . abstract construction of different types of VMs (we only care about Docker for now)
//   . manage lifecycle of VM (start with build, start, stop ...
//     eventually probably need fine grained management)
type VMController struct {
	sync.RWMutex
	// Handlers for each chaincode
	containerLocks map[string]*refCountedLock
}

//singleton...acess through NewVMController
var vmcontroller *VMController

//NewVMController - creates/returns singleton
func init() {
	vmcontroller = new(VMController)
	vmcontroller.containerLocks = make(map[string]*refCountedLock)
}

func (vmc *VMController) lockContainer(containerName string) {
	//get the container lock under global lock
	vmcontroller.Lock()
	var refLck *refCountedLock
	var ok bool
	if refLck, ok = vmcontroller.containerLocks[containerName]; !ok {
		refLck = &refCountedLock{refCount: 1, lock: &sync.RWMutex{}}
		vmcontroller.containerLocks[containerName] = refLck
	} else {
		refLck.refCount++
		vmLogger.Debugf("refcount %d (%s)", refLck.refCount, containerName)
	}
	vmcontroller.Unlock()
	vmLogger.Debugf("waiting for container(%s) lock", containerName)
	refLck.lock.Lock()
	vmLogger.Debugf("got container (%s) lock", containerName)
}

func (vmc *VMController) unlockContainer(containerName string) {
	vmcontroller.Lock()
	if refLck, ok := vmcontroller.containerLocks[containerName]; ok {
		if refLck.refCount <= 0 {
			panic("refcnt <= 0")
		}
		refLck.lock.Unlock()
		if refLck.refCount--; refLck.refCount == 0 {
			vmLogger.Debugf("container lock deleted(%s)", containerName)
			delete(vmcontroller.containerLocks, containerName)
		}
	} else {
		vmLogger.Debugf("no lock to unlock(%s)!!", containerName)
	}
	vmcontroller.Unlock()
}

//VMCReqIntf - all requests should implement this interface.
//The context should be passed and tested at each layer till we stop
//note that we'd stop on the first method on the stack that does not
//take context
type VMCReqIntf interface {
	do(ctxt context.Context, v api.VM) VMCResp
	getContainerName() string
}

//VMCResp - response from requests. resp field is a anon interface.
//It can hold any response. err should be tested first
type VMCResp struct {
	Err  error
	Resp interface{}
}

//CreateImageReq - properties for creating an container image
type CreateImageReq struct {
	ContainerName string
	Reader        io.Reader
	Args          []string
	Env           []string
}

func (bp CreateImageReq) do(ctxt context.Context, v api.VM) VMCResp {
	var resp VMCResp

	if err := v.Deploy(ctxt, bp.ContainerName, bp.Args, bp.Env, bp.Reader); err != nil {
		resp = VMCResp{Err: err}
	} else {
		resp = VMCResp{}
	}

	return resp
}

func (bp CreateImageReq) getContainerName() string {
	return bp.ContainerName
}

//StartImageReq - properties for starting a container.
type StartImageReq struct {
	ContainerName string
	Builder       api.BuildSpecFactory
	Args          []string
	Env           []string
	PrelaunchFunc api.PrelaunchFunc
}

func (si StartImageReq) do(ctxt context.Context, v api.VM) VMCResp {
	var resp VMCResp

	if err := v.Start(ctxt, si.ContainerName, si.Args, si.Env, si.Builder, si.PrelaunchFunc); err != nil {
		resp = VMCResp{Err: err}
	} else {
		resp = VMCResp{}
	}

	return resp
}

func (si StartImageReq) getContainerName() string {
	return si.ContainerName
}

//StopImageReq - properties for stopping a container.
type StopImageReq struct {
	ContainerName  string
	Timeout        uint
	Dontkill       bool //by default we will kill the container after stopping
	Dontremove     bool //by default we will remove the container after killing
	DonDeleteImage bool
}

func (si StopImageReq) do(ctxt context.Context, v api.VM) VMCResp {
	var resp VMCResp

	if err := v.Stop(ctxt, si.ContainerName, si.Timeout, si.Dontkill, si.Dontremove, si.DonDeleteImage); err != nil {
		resp = VMCResp{Err: err}
	} else {
		resp = VMCResp{}
	}

	return resp
}

func (si StopImageReq) getContainerName() string {
	return si.ContainerName
}

//DestroyImageReq - properties for stopping a container.
type DestroyImageReq struct {
	ContainerName string
	Timeout       uint
	Force         bool
	NoPrune       bool
}

func (di DestroyImageReq) do(ctxt context.Context, v api.VM) VMCResp {
	var resp VMCResp

	if err := v.Destroy(ctxt, di.ContainerName, di.Force, di.NoPrune); err != nil {
		resp = VMCResp{Err: err}
	} else {
		resp = VMCResp{}
	}

	return resp
}

func (di DestroyImageReq) getContainerName() string {
	return di.ContainerName
}

//VMCProcess should be used as follows
//   . construct a context
//   . construct req of the right type (e.g., CreateImageReq)
//   . call it in a go routine
//   . process response in the go routing
//context can be cancelled. VMCProcess will try to cancel calling functions if it can
//For instance docker clients api's such as BuildImage are not cancelable.
//In all cases VMCProcess will wait for the called go routine to return
func VMCProcess(ctxt context.Context, req VMCReqIntf) (interface{}, error) {
	v := dockercontroller.NewDockerVM()
	if v == nil {
		return nil, fmt.Errorf("Unknown VM type Docker")
	}
	c := make(chan struct{})
	var resp interface{}
	go func() {
		defer close(c)
		vmcontroller.lockContainer(req.getContainerName())
		resp = req.do(ctxt, v)
		vmcontroller.unlockContainer(req.getContainerName())
	}()

	select {
	case <-c:
		return resp, nil
	case <-ctxt.Done():
		//TODO cancel req.do ... (needed) ?
		<-c
		return nil, ctxt.Err()
	}
}
