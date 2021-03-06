package assigner

import (
	"encoding/json"
	"os/exec"
	"strconv"
	"time"

	"github.com/TTK4145-2022-students/driver-go-group-78/elevio"
	"github.com/TTK4145-2022-students/project-group-78/central"
	"github.com/TTK4145-2022-students/project-group-78/config"
	"github.com/TTK4145-2022-students/project-group-78/elevator"
)

// Assign returns orders assigned to cs.Origin (this elevator), after removing all faulty elevators.
// A faulty elevator is one that has orders assigned to it and has not moved in a while.
func Assign(cs central.CentralState) elevator.Orders {
	hrai := newHraInput(cs)
	for c := time.Duration(1); ; c++ {
		// Increase timeout for each removed elevator, so that not multiple elevators times out the same order.
		// The reason for finding anOTHER faulty elevator is that we want "ourselves" to try to serve orders.
		key, faulty := findOtherFaultyElevator(cs, hrai, c*config.OrderTimout, c*config.ElevTimeout)
		if faulty {
			delete(hrai.States, key)
		} else {
			return hallRequestAssigner(hrai)[strconv.Itoa(cs.Origin)]
		}
	}
}

type hraInput struct {
	HallRequests [config.NumFloors][config.NumHallOrderTypes]bool `json:"hallRequests"`
	States       map[string]hraState                              `json:"states"`
}

type hraState struct {
	Behaviour   string                 `json:"behaviour"`
	Floor       int                    `json:"floor"`
	Direction   string                 `json:"direction"`
	CabRequests [config.NumFloors]bool `json:"cabRequests"`
}

func newHraInput(cs central.CentralState) hraInput {
	hrai := hraInput{}
	for f := range cs.HallOrders {
		hrai.HallRequests[f] = [...]bool{cs.HallOrders[f][elevio.BT_HallUp].Active, cs.HallOrders[f][elevio.BT_HallDown].Active}
	}
	hrai.States = make(map[string]hraState)
	for id, state := range cs.States {
		hrai.States[strconv.Itoa(id)] = hraState{
			Behaviour:   state.Behaviour.ToString(),
			Floor:       state.Floor,
			Direction:   state.Direction.ToString(),
			CabRequests: cs.CabOrders[id],
		}
	}
	return hrai
}

func hallRequestAssigner(hrai hraInput) map[string]elevator.Orders {
	b, err := json.Marshal(hrai)
	if err != nil {
		panic(err)
	}

	output, err := exec.Command("./hall_request_assigner", "-i", "--includeCab", string(b)).Output()
	if err != nil {
		panic(err)
	}

	orders := make(map[string]elevator.Orders)
	err = json.Unmarshal(output, &orders)
	if err != nil {
		panic(err)
	}
	return orders
}

func findOtherFaultyElevator(cs central.CentralState, hrai hraInput, orderTimeout time.Duration, elevatorTimeout time.Duration) (key string, faulty bool) {
	for key, orders := range hallRequestAssigner(hrai) {
		id, err := strconv.Atoi(key)
		if err != nil {
			panic(err)
		}
		if id == cs.Origin {
			continue
		}
		for f := range orders {
			for bt := range orders[f] {
				if orders[f][bt] &&
					bt != elevio.BT_Cab &&
					time.Since(cs.HallOrders[f][bt].Time) > orderTimeout &&
					time.Since(cs.LastStateUpdate[id]) > elevatorTimeout {
					// If we have an old hall order and the assigned elevator has not responeded, conclude that it is faulty
					return key, true
				}
			}
		}
	}
	return
}
