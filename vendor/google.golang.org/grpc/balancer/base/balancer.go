/*
 *
 * Copyright 2017 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package base

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/resolver"
)

var logger = grpclog.Component("balancer")

type baseBuilder struct {
	name          string
	pickerBuilder PickerBuilder
	config        Config
}

func (bb *baseBuilder) Build(cc balancer.ClientConn, opt balancer.BuildOptions) balancer.Balancer {
	bal := &baseBalancer{
		cc:            cc,
		pickerBuilder: bb.pickerBuilder,

<<<<<<< HEAD
<<<<<<< HEAD
		subConns: resolver.NewAddressMap(),
=======
		subConns: make(map[resolver.Address]subConnInfo),
>>>>>>> upgrade to k8s 1.23 lib
=======
		subConns: resolver.NewAddressMap(),
>>>>>>> chore: Merge changes from upstream as of 2022-01-26 (#351)
		scStates: make(map[balancer.SubConn]connectivity.State),
		csEvltr:  &balancer.ConnectivityStateEvaluator{},
		config:   bb.config,
	}
	// Initialize picker to a picker that always returns
	// ErrNoSubConnAvailable, because when state of a SubConn changes, we
	// may call UpdateState with this picker.
	bal.picker = NewErrPicker(balancer.ErrNoSubConnAvailable)
	return bal
}

func (bb *baseBuilder) Name() string {
	return bb.name
}

<<<<<<< HEAD
<<<<<<< HEAD
=======
type subConnInfo struct {
	subConn balancer.SubConn
	attrs   *attributes.Attributes
}

>>>>>>> upgrade to k8s 1.23 lib
=======
>>>>>>> chore: Merge changes from upstream as of 2022-01-26 (#351)
type baseBalancer struct {
	cc            balancer.ClientConn
	pickerBuilder PickerBuilder

	csEvltr *balancer.ConnectivityStateEvaluator
	state   connectivity.State

<<<<<<< HEAD
<<<<<<< HEAD
	subConns *resolver.AddressMap
=======
	subConns map[resolver.Address]subConnInfo // `attributes` is stripped from the keys of this map (the addresses)
>>>>>>> upgrade to k8s 1.23 lib
=======
	subConns *resolver.AddressMap
>>>>>>> chore: Merge changes from upstream as of 2022-01-26 (#351)
	scStates map[balancer.SubConn]connectivity.State
	picker   balancer.Picker
	config   Config

	resolverErr error // the last error reported by the resolver; cleared on successful resolution
	connErr     error // the last connection error; cleared upon leaving TransientFailure
}

func (b *baseBalancer) ResolverError(err error) {
	b.resolverErr = err
<<<<<<< HEAD
<<<<<<< HEAD
	if b.subConns.Len() == 0 {
		b.state = connectivity.TransientFailure
	}

	if b.state != connectivity.TransientFailure {
		// The picker will not change since the balancer does not currently
		// report an error.
		return
	}
=======
	if len(b.subConns) == 0 {
=======
	if b.subConns.Len() == 0 {
>>>>>>> chore: Merge changes from upstream as of 2022-01-26 (#351)
		b.state = connectivity.TransientFailure
	}

	if b.state != connectivity.TransientFailure {
		// The picker will not change since the balancer does not currently
		// report an error.
		return
	}
>>>>>>> upgrade to k8s 1.23 lib
	b.regeneratePicker()
	b.cc.UpdateState(balancer.State{
		ConnectivityState: b.state,
		Picker:            b.picker,
	})
}

func (b *baseBalancer) UpdateClientConnState(s balancer.ClientConnState) error {
	// TODO: handle s.ResolverState.ServiceConfig?
	if logger.V(2) {
		logger.Info("base.baseBalancer: got new ClientConn state: ", s)
	}
	// Successful resolution; clear resolver error and ensure we return nil.
	b.resolverErr = nil
	// addrsSet is the set converted from addrs, it's used for quick lookup of an address.
	addrsSet := resolver.NewAddressMap()
	for _, a := range s.ResolverState.Addresses {
<<<<<<< HEAD
<<<<<<< HEAD
		addrsSet.Set(a, nil)
		if _, ok := b.subConns.Get(a); !ok {
=======
		// Strip attributes from addresses before using them as map keys. So
		// that when two addresses only differ in attributes pointers (but with
		// the same attribute content), they are considered the same address.
		//
		// Note that this doesn't handle the case where the attribute content is
		// different. So if users want to set different attributes to create
		// duplicate connections to the same backend, it doesn't work. This is
		// fine for now, because duplicate is done by setting Metadata today.
		//
		// TODO: read attributes to handle duplicate connections.
		aNoAttrs := a
		aNoAttrs.Attributes = nil
		addrsSet[aNoAttrs] = struct{}{}
		if scInfo, ok := b.subConns[aNoAttrs]; !ok {
>>>>>>> upgrade to k8s 1.23 lib
=======
		addrsSet.Set(a, nil)
		if _, ok := b.subConns.Get(a); !ok {
>>>>>>> chore: Merge changes from upstream as of 2022-01-26 (#351)
			// a is a new address (not existing in b.subConns).
			sc, err := b.cc.NewSubConn([]resolver.Address{a}, balancer.NewSubConnOptions{HealthCheckEnabled: b.config.HealthCheck})
			if err != nil {
				logger.Warningf("base.baseBalancer: failed to create new SubConn: %v", err)
				continue
			}
<<<<<<< HEAD
<<<<<<< HEAD
			b.subConns.Set(a, sc)
=======
			b.subConns[aNoAttrs] = subConnInfo{subConn: sc, attrs: a.Attributes}
>>>>>>> upgrade to k8s 1.23 lib
=======
			b.subConns.Set(a, sc)
>>>>>>> chore: Merge changes from upstream as of 2022-01-26 (#351)
			b.scStates[sc] = connectivity.Idle
			b.csEvltr.RecordTransition(connectivity.Shutdown, connectivity.Idle)
			sc.Connect()
		}
	}
<<<<<<< HEAD
<<<<<<< HEAD
	for _, a := range b.subConns.Keys() {
		sci, _ := b.subConns.Get(a)
		sc := sci.(balancer.SubConn)
		// a was removed by resolver.
		if _, ok := addrsSet.Get(a); !ok {
			b.cc.RemoveSubConn(sc)
			b.subConns.Delete(a)
=======
	for a, scInfo := range b.subConns {
		// a was removed by resolver.
		if _, ok := addrsSet[a]; !ok {
			b.cc.RemoveSubConn(scInfo.subConn)
			delete(b.subConns, a)
>>>>>>> upgrade to k8s 1.23 lib
=======
	for _, a := range b.subConns.Keys() {
		sci, _ := b.subConns.Get(a)
		sc := sci.(balancer.SubConn)
		// a was removed by resolver.
		if _, ok := addrsSet.Get(a); !ok {
			b.cc.RemoveSubConn(sc)
			b.subConns.Delete(a)
>>>>>>> chore: Merge changes from upstream as of 2022-01-26 (#351)
			// Keep the state of this sc in b.scStates until sc's state becomes Shutdown.
			// The entry will be deleted in UpdateSubConnState.
		}
	}
	// If resolver state contains no addresses, return an error so ClientConn
	// will trigger re-resolve. Also records this as an resolver error, so when
	// the overall state turns transient failure, the error message will have
	// the zero address information.
	if len(s.ResolverState.Addresses) == 0 {
		b.ResolverError(errors.New("produced zero addresses"))
		return balancer.ErrBadResolverState
	}
	return nil
}

// mergeErrors builds an error from the last connection error and the last
// resolver error.  Must only be called if b.state is TransientFailure.
func (b *baseBalancer) mergeErrors() error {
	// connErr must always be non-nil unless there are no SubConns, in which
	// case resolverErr must be non-nil.
	if b.connErr == nil {
		return fmt.Errorf("last resolver error: %v", b.resolverErr)
	}
	if b.resolverErr == nil {
		return fmt.Errorf("last connection error: %v", b.connErr)
	}
	return fmt.Errorf("last connection error: %v; last resolver error: %v", b.connErr, b.resolverErr)
}

// regeneratePicker takes a snapshot of the balancer, and generates a picker
// from it. The picker is
//  - errPicker if the balancer is in TransientFailure,
//  - built by the pickerBuilder with all READY SubConns otherwise.
func (b *baseBalancer) regeneratePicker() {
	if b.state == connectivity.TransientFailure {
		b.picker = NewErrPicker(b.mergeErrors())
		return
	}
	readySCs := make(map[balancer.SubConn]SubConnInfo)

	// Filter out all ready SCs from full subConn map.
<<<<<<< HEAD
<<<<<<< HEAD
=======
>>>>>>> chore: Merge changes from upstream as of 2022-01-26 (#351)
	for _, addr := range b.subConns.Keys() {
		sci, _ := b.subConns.Get(addr)
		sc := sci.(balancer.SubConn)
		if st, ok := b.scStates[sc]; ok && st == connectivity.Ready {
			readySCs[sc] = SubConnInfo{Address: addr}
<<<<<<< HEAD
=======
	for addr, scInfo := range b.subConns {
		if st, ok := b.scStates[scInfo.subConn]; ok && st == connectivity.Ready {
			addr.Attributes = scInfo.attrs
			readySCs[scInfo.subConn] = SubConnInfo{Address: addr}
>>>>>>> upgrade to k8s 1.23 lib
=======
>>>>>>> chore: Merge changes from upstream as of 2022-01-26 (#351)
		}
	}
	b.picker = b.pickerBuilder.Build(PickerBuildInfo{ReadySCs: readySCs})
}

func (b *baseBalancer) UpdateSubConnState(sc balancer.SubConn, state balancer.SubConnState) {
	s := state.ConnectivityState
	if logger.V(2) {
		logger.Infof("base.baseBalancer: handle SubConn state change: %p, %v", sc, s)
	}
	oldS, ok := b.scStates[sc]
	if !ok {
		if logger.V(2) {
			logger.Infof("base.baseBalancer: got state changes for an unknown SubConn: %p, %v", sc, s)
<<<<<<< HEAD
		}
		return
	}
	if oldS == connectivity.TransientFailure &&
		(s == connectivity.Connecting || s == connectivity.Idle) {
		// Once a subconn enters TRANSIENT_FAILURE, ignore subsequent IDLE or
		// CONNECTING transitions to prevent the aggregated state from being
		// always CONNECTING when many backends exist but are all down.
		if s == connectivity.Idle {
			sc.Connect()
=======
>>>>>>> upgrade to k8s 1.23 lib
		}
		return
	}
	if oldS == connectivity.TransientFailure && s == connectivity.Connecting {
		// Once a subconn enters TRANSIENT_FAILURE, ignore subsequent
		// CONNECTING transitions to prevent the aggregated state from being
		// always CONNECTING when many backends exist but are all down.
		return
	}
	b.scStates[sc] = s
	switch s {
	case connectivity.Idle:
		sc.Connect()
	case connectivity.Shutdown:
		// When an address was removed by resolver, b called RemoveSubConn but
		// kept the sc's state in scStates. Remove state for this sc here.
		delete(b.scStates, sc)
	case connectivity.TransientFailure:
		// Save error to be reported via picker.
		b.connErr = state.ConnectionError
	}

	b.state = b.csEvltr.RecordTransition(oldS, s)

	// Regenerate picker when one of the following happens:
	//  - this sc entered or left ready
	//  - the aggregated state of balancer is TransientFailure
	//    (may need to update error message)
	if (s == connectivity.Ready) != (oldS == connectivity.Ready) ||
		b.state == connectivity.TransientFailure {
		b.regeneratePicker()
	}
<<<<<<< HEAD
=======

>>>>>>> upgrade to k8s 1.23 lib
	b.cc.UpdateState(balancer.State{ConnectivityState: b.state, Picker: b.picker})
}

// Close is a nop because base balancer doesn't have internal state to clean up,
// and it doesn't need to call RemoveSubConn for the SubConns.
func (b *baseBalancer) Close() {
}

<<<<<<< HEAD
// ExitIdle is a nop because the base balancer attempts to stay connected to
// all SubConns at all times.
func (b *baseBalancer) ExitIdle() {
}

=======
>>>>>>> upgrade to k8s 1.23 lib
// NewErrPicker returns a Picker that always returns err on Pick().
func NewErrPicker(err error) balancer.Picker {
	return &errPicker{err: err}
}

// NewErrPickerV2 is temporarily defined for backward compatibility reasons.
//
// Deprecated: use NewErrPicker instead.
var NewErrPickerV2 = NewErrPicker

type errPicker struct {
	err error // Pick() always returns this err.
}

func (p *errPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	return balancer.PickResult{}, p.err
}
