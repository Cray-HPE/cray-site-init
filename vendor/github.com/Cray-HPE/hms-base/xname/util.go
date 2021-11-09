package xname

import (
	"errors"

	base "github.com/Cray-HPE/hms-base"
)

var ErrUnknownStruct = errors.New("unable to determine HMS Type from struct")

func GetHMSType(obj interface{}) (base.HMSType, error) {
	// Handy bash fragment to generate the type switch below
	// for hms_type in $(cat ./xname/types.go | grep '^type' | awk '{print $2}'); do
	// echo "	case $hms_type, *$hms_type:"
	// echo "		return base.$hms_type, nil"
	// done
	switch obj.(type) {
	case CDU, *CDU:
		return base.CDU, nil
	case CDUMgmtSwitch, *CDUMgmtSwitch:
		return base.CDUMgmtSwitch, nil
	case Cabinet, *Cabinet:
		return base.Cabinet, nil
	case CabinetPDUController, *CabinetPDUController:
		return base.CabinetPDUController, nil
	case Chassis, *Chassis:
		return base.Chassis, nil
	case MgmtSwitch, *MgmtSwitch:
		return base.MgmtSwitch, nil
	case MgmtSwitchConnector, *MgmtSwitchConnector:
		return base.MgmtSwitchConnector, nil
	case MgmtHLSwitchEnclosure, *MgmtHLSwitchEnclosure:
		return base.MgmtHLSwitchEnclosure, nil
	case MgmtHLSwitch, *MgmtHLSwitch:
		return base.MgmtHLSwitch, nil
	case RouterModule, *RouterModule:
		return base.RouterModule, nil
	case RouterBMC, *RouterBMC:
		return base.RouterBMC, nil
	case ComputeModule, *ComputeModule:
		return base.ComputeModule, nil
	case NodeBMC, *NodeBMC:
		return base.NodeBMC, nil
	case Node, *Node:
		return base.Node, nil
	}

	return base.HMSTypeInvalid, ErrUnknownStruct
}
