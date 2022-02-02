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

	"github.com/openconfig/ondatra/fakebind/fakedevice"
	"github.com/openconfig/ondatra/internal/binding"
	"github.com/openconfig/ondatra/internal/reservation"
	"github.com/pborman/uuid"
	"google.golang.org/grpc"

	"github.com/openconfig/gribigo/gnmit"

	gpb "github.com/openconfig/gnmi/proto/gnmi"
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
		c, addr, err := fakedevice.NewTarget(ctx, fmt.Sprintf("localhost:%d", i), dut.GetId())
		if err != nil {
			return nil, err
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
