// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package fakebind provides an Ondatra binding for fake devices.
package fakebind

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	log "github.com/golang/glog"
	"github.com/openconfig/ondatra/internal/binding"
	"github.com/openconfig/ondatra/internal/reservation"
	"github.com/openconfig/ondatra/telemetry"
	"github.com/openconfig/ondatra/telemetry/device"
	"github.com/openconfig/ygot/ygot"
	"github.com/pborman/uuid"
	"google.golang.org/grpc"

	"github.com/openconfig/gribigo/gnmit"

	"github.com/openconfig/gnmi/coalesce"
	"github.com/openconfig/gnmi/ctree"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/gnmi/value"
	opb "github.com/openconfig/ondatra/proto"
)

type fakeDUT struct {
	target *gnmit.Collector
	addr   string
}

// Bind implements the ondatra Binding interface for KNE
type Bind struct {
	binding.Binding
	// fakeDUTs contains references to all the instantiated fake DUTs.
	fakeDUTs map[string]*fakeDUT
}

// New returns a new KNE bind instance.
func New() (*Bind, error) {
	return &Bind{
		fakeDUTs: map[string]*fakeDUT{},
	}, nil
}

// bootTimeTask is a task that updates the boot-time leaf with the current
// time. It does not spawn any long-running threads.
func bootTimeTask(_ gnmit.Queue, update gnmit.UpdateFn, target string, remove func()) error {
	defer remove()
	p0, _, errs := ygot.ResolvePath(device.DeviceRoot("").System().BootTime())
	if errs != nil {
		return fmt.Errorf("bootTimeTask failed to initialize due to error: %v", errs)
	}

	now, err := value.FromScalar(time.Now().UnixNano())
	if err != nil {
		return fmt.Errorf("bootTimeTask: %v", err)
	}
	log.V(2).Infof("bootTimeTask: %v, %v", p0, now)
	if err := update(&gpb.Notification{
		Timestamp: time.Now().UnixNano(),
		Prefix: &gpb.Path{
			Origin: "openconfig",
			Target: target,
		},
		Update: []*gpb.Update{{
			Path: p0,
			Val:  now,
		}},
	}); err != nil {
		return err
	}

	return nil
}

// currentDateTimeTask updates the current-datetime leaf with the current time,
// and spawns a thread that wakes up every second to update the leaf.
func currentDateTimeTask(_ gnmit.Queue, update gnmit.UpdateFn, target string, remove func()) error {
	p0, _, err := ygot.ResolvePath(device.DeviceRoot("").System().CurrentDatetime())
	if err != nil {
		return fmt.Errorf("currentDateTimeTask failed to initialize due to error: %v", err)
	}

	tick := time.Tick(time.Second)
	if tick == nil {
		return fmt.Errorf("currentDateTimeTask: tick is nil")
	}

	periodic := func() error {
		currentDatetime, err := value.FromScalar(time.Now().Format(time.RFC3339))
		if err != nil {
			return fmt.Errorf("currentDateTimeTask: %v", err)
		}
		log.V(2).Infof("currentDateTimeTask: %v, %v", p0, currentDatetime)
		if err := update(&gpb.Notification{
			Timestamp: time.Now().UnixNano(),
			Prefix: &gpb.Path{
				Origin: "openconfig",
				Target: target,
			},
			Update: []*gpb.Update{{
				Path: p0,
				Val:  currentDatetime,
			}},
		}); err != nil {
			return err
		}
		return nil
	}

	if err := periodic(); err != nil {
		return err
	}

	go func() {
		defer remove()
		for range tick {
			if err := periodic(); err != nil {
				log.Errorf("currentDateTimeTask error: %v", err)
				return
			}
		}
	}()

	return nil
}

// syslogTask is a meaningless test task that monitors updates to the
// current-datetime leaf and writes updates to the syslog message leaf whenever
// the current-datetime leaf is updated.
func syslogTask(q gnmit.Queue, update gnmit.UpdateFn, target string, remove func()) error {
	p0, _, err := ygot.ResolvePath(device.DeviceRoot("").System().Messages().Message().Msg())
	if err != nil {
		log.Errorf("syslogTask failed to initialize due to error: %v", err)
	}

	go func() {
		defer remove()
		for {
			item, _, err := q.Next(context.Background())
			if coalesce.IsClosedQueue(err) {
				return
			}
			n, ok := item.(*ctree.Leaf)
			if !ok || n == nil {
				log.Errorf("syslogTask invalid cache node: %#v", item)
				return
			}
			v := n.Value()
			no, ok := v.(*gpb.Notification)
			if !ok || no == nil {
				log.Errorf("syslogTask invalid cache node, expected non-nil *gpb.Notification type, got: %#v", v)
				return
			}
			for _, u := range no.Update {
				sv, err := value.ToScalar(u.Val)
				if err != nil {
					log.Errorf("syslogTask: %v", err)
					return
				}
				strv, ok := sv.(string)
				if !ok {
					log.Errorf("syslogTask: cannot convert to string, got (%T, %v)", sv, sv)
					return
				}
				syslog, err := value.FromScalar("current date-time updated to " + strv)
				if err != nil {
					log.Errorf("syslogTask: %v", err)
					return
				}
				if err := update(&gpb.Notification{
					Timestamp: time.Now().UnixNano(),
					Prefix: &gpb.Path{
						Origin: "openconfig",
						Target: target,
					},
					Update: []*gpb.Update{{
						Path: p0,
						Val:  syslog,
					}},
				}); err != nil {
					log.Errorf("syslogTask: %v", err)
					return
				}
			}
			for _, _ = range no.Delete {
			}
		}
	}()

	return nil
}

// tasks returns the set of functions that should be called that may read
// and/or modify internal state.
//
// These tasks are invoked during the creation of each device's Subscribe
// server and may spawn long-running or future-running thread(s) that make
// modifications to a device's cache.
func tasks(target string) []gnmit.Task {
	p0, _, err := ygot.ResolvePath(device.DeviceRoot("").System().CurrentDatetime())
	if err != nil {
		panic(fmt.Sprintf("currentDateTimeTask failed to initialize due to error: %v", err))
	}

	return []gnmit.Task{{
		Run: currentDateTimeTask,
		// No paths means the task should periodically wake up itself.
		Paths: []*gpb.Path{},
		Prefix: &gpb.Path{
			Origin: "openconfig",
			Target: target,
		},
	}, {
		Run: bootTimeTask,
		// No paths means the task should periodically wake up itself.
		Paths: []*gpb.Path{},
		Prefix: &gpb.Path{
			Origin: "openconfig",
			Target: target,
		},
	}, {
		Run: syslogTask,
		// No paths means the task should periodically wake up itself.
		Paths: []*gpb.Path{
			p0,
		},
		Prefix: &gpb.Path{
			Origin: "openconfig",
			Target: target,
		},
	}}
}

// Reserve implements the binding Reserve method by finding nodes and links in
// the topology specified in the config file that match the requested testbed.
func (b *Bind) Reserve(ctx context.Context, tb *opb.Testbed, runTime time.Duration, waitTime time.Duration) (*reservation.Reservation, error) {
	res := &reservation.Reservation{
		ID:   uuid.New(),
		DUTs: make(map[string]*reservation.DUT),
		ATEs: make(map[string]*reservation.ATE),
	}
	for i, dut := range tb.GetDuts() {
		if i > 0 {
			// TODO(wenbli): Support multiple fake devices.
			return nil, fmt.Errorf("can only support a single dut for now")
		}
		ports := make(map[string]*reservation.Port)
		for _, port := range dut.GetPorts() {
			// Just give the port a dummy name to pass validation.
			ports[port.Id] = &reservation.Port{Name: port.Id}
		}
		res.DUTs[dut.GetId()] = &reservation.DUT{
			Dims: &reservation.Dims{
				Name:  dut.GetId(),
				Ports: ports,
			},
		}
		schema, err := telemetry.Schema()
		if err != nil {
			return nil, err
		}
		c, addr, err := gnmit.NewSettable(ctx, fmt.Sprintf("localhost:%d", i), dut.GetId(), false, schema, tasks(dut.GetId()))
		if err != nil {
			return nil, fmt.Errorf("cannot start server, got err: %v", err)
		}
		b.fakeDUTs[dut.GetId()] = &fakeDUT{
			target: c,
			addr:   addr,
		}
	}
	for _, _ = range tb.GetAtes() {
		return nil, fmt.Errorf("ATEs currently not supported by fakebind")
	}
	return res, nil
}

// Release is a no-op because there's need to reserve local VMs.
func (b *Bind) Release(_ context.Context) error {
	for _, fake := range b.fakeDUTs {
		fake.target.Stop()
	}
	return nil
}

// SetTestMetadata is unused for KNE.
func (b *Bind) SetTestMetadata(_ *binding.TestMetadata) error {
	return nil
}

func (b *Bind) DialGNMI(ctx context.Context, dut *reservation.DUT, opts ...grpc.DialOption) (gpb.GNMIClient, error) {
	conn, err := grpc.Dial(b.fakeDUTs[dut.Name].addr, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("cannot dial gNMI server, %v", err)
	}

	return gpb.NewGNMIClient(conn), nil
}
