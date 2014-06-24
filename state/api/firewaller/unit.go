// Copyright 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package firewaller

import (
	"fmt"

	"github.com/juju/names"

	"github.com/juju/juju/network"
	"github.com/juju/juju/state/api/common"
	"github.com/juju/juju/state/api/params"
	"github.com/juju/juju/state/api/watcher"
)

// Unit represents a juju unit as seen by a firewaller worker.
type Unit struct {
	st   *State
	tag  string
	life params.Life
}

// Name returns the name of the unit.
func (u *Unit) Name() string {
	return mustParseUnitTag(u.tag).Id()
}

func mustParseUnitTag(unitTag string) names.UnitTag {
	tag, err := names.ParseUnitTag(unitTag)
	if err != nil {
		panic(err)
	}
	return tag
}

// Life returns the unit's life cycle value.
func (u *Unit) Life() params.Life {
	return u.life
}

// Refresh updates the cached local copy of the unit's data.
func (u *Unit) Refresh() error {
	life, err := common.Life(u.st.caller, u.tag)
	if err != nil {
		return err
	}
	u.life = life
	return nil
}

// Watch returns a watcher for observing changes to the unit.
func (u *Unit) Watch() (watcher.NotifyWatcher, error) {
	var results params.NotifyWatchResults
	args := params.Entities{
		Entities: []params.Entity{{Tag: u.tag}},
	}
	err := u.st.caller.FacadeCall("Watch", args, &results)
	if err != nil {
		return nil, err
	}
	if len(results.Results) != 1 {
		return nil, fmt.Errorf("expected 1 result, got %d", len(results.Results))
	}
	result := results.Results[0]
	if result.Error != nil {
		return nil, result.Error
	}
	w := watcher.NewNotifyWatcher(u.st.caller.RawAPICaller(), result)
	return w, nil
}

// Service returns the service.
func (u *Unit) Service() (*Service, error) {
	serviceTag := names.NewServiceTag(names.UnitService(u.Name()))
	service := &Service{
		st:  u.st,
		tag: serviceTag.String(),
	}
	// Call Refresh() immediately to get the up-to-date
	// life and other needed locally cached fields.
	err := service.Refresh()
	if err != nil {
		return nil, err
	}
	return service, nil
}

// OpenedPorts returns the list of opened ports for this unit.
//
// NOTE: This differs from state.Unit.OpenedPorts() by returning
// an error as well, because it needs to make an API call.
func (u *Unit) OpenedPorts() ([]network.Port, error) {
	var results params.PortsResults
	args := params.Entities{
		Entities: []params.Entity{{Tag: u.tag}},
	}
	err := u.st.caller.FacadeCall("OpenedPorts", args, &results)
	if err != nil {
		return nil, err
	}
	if len(results.Results) != 1 {
		return nil, fmt.Errorf("expected 1 result, got %d", len(results.Results))
	}
	result := results.Results[0]
	if result.Error != nil {
		return nil, result.Error
	}
	return result.Ports, nil
}

// AssignedMachine returns the tag of this unit's assigned machine (if
// any), or a CodeNotAssigned error.
func (u *Unit) AssignedMachine() (string, error) {
	var results params.StringResults
	args := params.Entities{
		Entities: []params.Entity{{Tag: u.tag}},
	}
	err := u.st.caller.FacadeCall("GetAssignedMachine", args, &results)
	if err != nil {
		return "", err
	}
	if len(results.Results) != 1 {
		return "", fmt.Errorf("expected 1 result, got %d", len(results.Results))
	}
	result := results.Results[0]
	if result.Error != nil {
		return "", result.Error
	}
	return result.Result, nil
}
