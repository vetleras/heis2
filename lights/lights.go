package lights

import (
	"github.com/TTK4145-2022-students/driver-go-group-78/elevio"
	"github.com/TTK4145-2022-students/project-group-78/central"
	"github.com/TTK4145-2022-students/project-group-78/config"
	"github.com/TTK4145-2022-students/project-group-78/elevator"
)

var lights elevator.Orders

func Set(cs central.CentralState) {
	localCabOrders := cs.CabOrders[cs.Origin]
	for f := range localCabOrders {
		if localCabOrders[f] != lights[f][elevio.BT_Cab] {
			elevio.SetButtonLamp(elevio.BT_Cab, f, localCabOrders[f])
			lights[f][elevio.BT_Cab] = localCabOrders[f]
		}
	}
	for f := range cs.HallOrders {
		for bt := range cs.HallOrders[f] {
			value := cs.HallOrders[f][bt].Active
			if value != lights[f][bt] {
				elevio.SetButtonLamp(elevio.ButtonType(bt), f, value)
				lights[f][bt] = value
			}
		}
	}
}

func Clear() {
	for f := 0; f < config.NumFloors; f++ {
		for bt := 0; bt < config.NumOrderTypes; bt++ {
			elevio.SetButtonLamp(elevio.ButtonType(bt), f, false)
		}
	}
}
