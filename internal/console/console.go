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

// Package console provides implementation on connecting to a DUT's console.
package console

import (
	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"github.com/openconfig/ondatra/binding"
	"github.com/openconfig/ondatra/internal/testbed"
)

// NewConsole creates a console client for the specified DUT.
func NewConsole(ctx context.Context, dut *binding.DUT) (binding.StreamClient, error) {
	return testbed.Bind().DialConsole(ctx, dut, grpc.WithBlock())
}
