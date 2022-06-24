// MIT License
//
// (C) Copyright [2021-2022] Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package xnames

import (
	"errors"
	"strconv"

	"github.com/Cray-HPE/hms-xname/xnametypes"
)

var ErrUnknownStruct = errors.New("unable to determine HMS Type from struct")

// GetHMSType for a given xname structure will return its HMSType
// If the given object is not a structure from the xnames package,
// then the ErrUnknownStruct will be returned along with HMSTypeInvalid
func GetHMSType(obj interface{}) (xnametypes.HMSType, error) {
	switch obj.(type) {

	case System, *System:
		return xnametypes.System, nil
	case CDU, *CDU:
		return xnametypes.CDU, nil
	case CDUMgmtSwitch, *CDUMgmtSwitch:
		return xnametypes.CDUMgmtSwitch, nil
	case Cabinet, *Cabinet:
		return xnametypes.Cabinet, nil
	case CEC, *CEC:
		return xnametypes.CEC, nil
	case CabinetBMC, *CabinetBMC:
		return xnametypes.CabinetBMC, nil
	case CabinetCDU, *CabinetCDU:
		return xnametypes.CabinetCDU, nil
	case CabinetPDUController, *CabinetPDUController:
		return xnametypes.CabinetPDUController, nil
	case CabinetPDU, *CabinetPDU:
		return xnametypes.CabinetPDU, nil
	case CabinetPDUOutlet, *CabinetPDUOutlet:
		return xnametypes.CabinetPDUOutlet, nil
	case CabinetPDUPowerConnector, *CabinetPDUPowerConnector:
		return xnametypes.CabinetPDUPowerConnector, nil
	case CabinetPDUNic, *CabinetPDUNic:
		return xnametypes.CabinetPDUNic, nil
	case Chassis, *Chassis:
		return xnametypes.Chassis, nil
	case CMMFpga, *CMMFpga:
		return xnametypes.CMMFpga, nil
	case CMMRectifier, *CMMRectifier:
		return xnametypes.CMMRectifier, nil
	case ChassisBMC, *ChassisBMC:
		return xnametypes.ChassisBMC, nil
	case ChassisBMCNic, *ChassisBMCNic:
		return xnametypes.ChassisBMCNic, nil
	case ComputeModule, *ComputeModule:
		return xnametypes.ComputeModule, nil
	case NodeBMC, *NodeBMC:
		return xnametypes.NodeBMC, nil
	case Node, *Node:
		return xnametypes.Node, nil
	case Memory, *Memory:
		return xnametypes.Memory, nil
	case NodeAccel, *NodeAccel:
		return xnametypes.NodeAccel, nil
	case NodeAccelRiser, *NodeAccelRiser:
		return xnametypes.NodeAccelRiser, nil
	case NodeHsnNic, *NodeHsnNic:
		return xnametypes.NodeHsnNic, nil
	case NodeNic, *NodeNic:
		return xnametypes.NodeNic, nil
	case Processor, *Processor:
		return xnametypes.Processor, nil
	case StorageGroup, *StorageGroup:
		return xnametypes.StorageGroup, nil
	case Drive, *Drive:
		return xnametypes.Drive, nil
	case NodeBMCNic, *NodeBMCNic:
		return xnametypes.NodeBMCNic, nil
	case NodeEnclosure, *NodeEnclosure:
		return xnametypes.NodeEnclosure, nil
	case NodeEnclosurePowerSupply, *NodeEnclosurePowerSupply:
		return xnametypes.NodeEnclosurePowerSupply, nil
	case NodeFpga, *NodeFpga:
		return xnametypes.NodeFpga, nil
	case NodePowerConnector, *NodePowerConnector:
		return xnametypes.NodePowerConnector, nil
	case MgmtHLSwitchEnclosure, *MgmtHLSwitchEnclosure:
		return xnametypes.MgmtHLSwitchEnclosure, nil
	case MgmtHLSwitch, *MgmtHLSwitch:
		return xnametypes.MgmtHLSwitch, nil
	case MgmtSwitch, *MgmtSwitch:
		return xnametypes.MgmtSwitch, nil
	case MgmtSwitchConnector, *MgmtSwitchConnector:
		return xnametypes.MgmtSwitchConnector, nil
	case RouterModule, *RouterModule:
		return xnametypes.RouterModule, nil
	case HSNAsic, *HSNAsic:
		return xnametypes.HSNAsic, nil
	case HSNLink, *HSNLink:
		return xnametypes.HSNLink, nil
	case HSNBoard, *HSNBoard:
		return xnametypes.HSNBoard, nil
	case HSNConnector, *HSNConnector:
		return xnametypes.HSNConnector, nil
	case HSNConnectorPort, *HSNConnectorPort:
		return xnametypes.HSNConnectorPort, nil
	case RouterBMC, *RouterBMC:
		return xnametypes.RouterBMC, nil
	case RouterBMCNic, *RouterBMCNic:
		return xnametypes.RouterBMCNic, nil
	case RouterFpga, *RouterFpga:
		return xnametypes.RouterFpga, nil
	case RouterPowerConnector, *RouterPowerConnector:
		return xnametypes.RouterPowerConnector, nil
	case RouterTOR, *RouterTOR:
		return xnametypes.RouterTOR, nil
	case RouterTORFpga, *RouterTORFpga:
		return xnametypes.RouterTORFpga, nil
	}
	return xnametypes.HMSTypeInvalid, ErrUnknownStruct
}

// FromString will convert the string representation of a xname into a xname structure
// If the string is not a valid xname, then nil and HMSTypeInvalid will be returned.
func FromString(xname string) Xname {
	hmsType := xnametypes.GetHMSType(xname)
	if hmsType == xnametypes.HMSTypeInvalid {
		return nil
	}

	re, err := xnametypes.GetHMSTypeRegex(hmsType)
	if err != nil {
		return nil
	}

	_, argCount, err := xnametypes.GetHMSTypeFormatString(hmsType)
	if err != nil {
		return nil
	}

	matchesRaw := re.FindStringSubmatch(xname)
	if (argCount + 1) != len(matchesRaw) {
		return nil
	}

	// If we have gotten to this point these matches should be integers, so we can safely convert them
	// to integers from strings.
	matches := []int{}
	for _, matchRaw := range matchesRaw[1:] {
		match, err := strconv.Atoi(matchRaw)
		if err != nil {
			return nil
		}

		matches = append(matches, match)
	}

	var component Xname

	switch hmsType {
	case xnametypes.System:
		component = System{}
	case xnametypes.CDU:
		component = CDU{
			CDU: matches[0],
		}
	case xnametypes.CDUMgmtSwitch:
		component = CDUMgmtSwitch{
			CDU:           matches[0],
			CDUMgmtSwitch: matches[1],
		}
	case xnametypes.Cabinet:
		component = Cabinet{
			Cabinet: matches[0],
		}
	case xnametypes.CEC:
		component = CEC{
			Cabinet: matches[0],
			CEC:     matches[1],
		}
	case xnametypes.CabinetBMC:
		component = CabinetBMC{
			Cabinet:    matches[0],
			CabinetBMC: matches[1],
		}
	case xnametypes.CabinetCDU:
		component = CabinetCDU{
			Cabinet:    matches[0],
			CabinetCDU: matches[1],
		}
	case xnametypes.CabinetPDUController:
		component = CabinetPDUController{
			Cabinet:              matches[0],
			CabinetPDUController: matches[1],
		}
	case xnametypes.CabinetPDU:
		component = CabinetPDU{
			Cabinet:              matches[0],
			CabinetPDUController: matches[1],
			CabinetPDU:           matches[2],
		}
	case xnametypes.CabinetPDUOutlet:
		component = CabinetPDUOutlet{
			Cabinet:              matches[0],
			CabinetPDUController: matches[1],
			CabinetPDU:           matches[2],
			CabinetPDUOutlet:     matches[3],
		}
	case xnametypes.CabinetPDUPowerConnector:
		component = CabinetPDUPowerConnector{
			Cabinet:                  matches[0],
			CabinetPDUController:     matches[1],
			CabinetPDU:               matches[2],
			CabinetPDUPowerConnector: matches[3],
		}
	case xnametypes.CabinetPDUNic:
		component = CabinetPDUNic{
			Cabinet:              matches[0],
			CabinetPDUController: matches[1],
			CabinetPDUNic:        matches[2],
		}
	case xnametypes.Chassis:
		component = Chassis{
			Cabinet: matches[0],
			Chassis: matches[1],
		}
	case xnametypes.CMMFpga:
		component = CMMFpga{
			Cabinet: matches[0],
			Chassis: matches[1],
			CMMFpga: matches[2],
		}
	case xnametypes.CMMRectifier:
		component = CMMRectifier{
			Cabinet:      matches[0],
			Chassis:      matches[1],
			CMMRectifier: matches[2],
		}
	case xnametypes.ChassisBMC:
		component = ChassisBMC{
			Cabinet:    matches[0],
			Chassis:    matches[1],
			ChassisBMC: matches[2],
		}
	case xnametypes.ChassisBMCNic:
		component = ChassisBMCNic{
			Cabinet:       matches[0],
			Chassis:       matches[1],
			ChassisBMC:    matches[2],
			ChassisBMCNic: matches[3],
		}
	case xnametypes.ComputeModule:
		component = ComputeModule{
			Cabinet:       matches[0],
			Chassis:       matches[1],
			ComputeModule: matches[2],
		}
	case xnametypes.NodeBMC:
		component = NodeBMC{
			Cabinet:       matches[0],
			Chassis:       matches[1],
			ComputeModule: matches[2],
			NodeBMC:       matches[3],
		}
	case xnametypes.Node:
		component = Node{
			Cabinet:       matches[0],
			Chassis:       matches[1],
			ComputeModule: matches[2],
			NodeBMC:       matches[3],
			Node:          matches[4],
		}
	case xnametypes.Memory:
		component = Memory{
			Cabinet:       matches[0],
			Chassis:       matches[1],
			ComputeModule: matches[2],
			NodeBMC:       matches[3],
			Node:          matches[4],
			Memory:        matches[5],
		}
	case xnametypes.NodeAccel:
		component = NodeAccel{
			Cabinet:       matches[0],
			Chassis:       matches[1],
			ComputeModule: matches[2],
			NodeBMC:       matches[3],
			Node:          matches[4],
			NodeAccel:     matches[5],
		}
	case xnametypes.NodeAccelRiser:
		component = NodeAccelRiser{
			Cabinet:        matches[0],
			Chassis:        matches[1],
			ComputeModule:  matches[2],
			NodeBMC:        matches[3],
			Node:           matches[4],
			NodeAccelRiser: matches[5],
		}
	case xnametypes.NodeHsnNic:
		component = NodeHsnNic{
			Cabinet:       matches[0],
			Chassis:       matches[1],
			ComputeModule: matches[2],
			NodeBMC:       matches[3],
			Node:          matches[4],
			NodeHsnNic:    matches[5],
		}
	case xnametypes.NodeNic:
		component = NodeNic{
			Cabinet:       matches[0],
			Chassis:       matches[1],
			ComputeModule: matches[2],
			NodeBMC:       matches[3],
			Node:          matches[4],
			NodeNic:       matches[5],
		}
	case xnametypes.Processor:
		component = Processor{
			Cabinet:       matches[0],
			Chassis:       matches[1],
			ComputeModule: matches[2],
			NodeBMC:       matches[3],
			Node:          matches[4],
			Processor:     matches[5],
		}
	case xnametypes.StorageGroup:
		component = StorageGroup{
			Cabinet:       matches[0],
			Chassis:       matches[1],
			ComputeModule: matches[2],
			NodeBMC:       matches[3],
			Node:          matches[4],
			StorageGroup:  matches[5],
		}
	case xnametypes.Drive:
		component = Drive{
			Cabinet:       matches[0],
			Chassis:       matches[1],
			ComputeModule: matches[2],
			NodeBMC:       matches[3],
			Node:          matches[4],
			StorageGroup:  matches[5],
			Drive:         matches[6],
		}
	case xnametypes.NodeBMCNic:
		component = NodeBMCNic{
			Cabinet:       matches[0],
			Chassis:       matches[1],
			ComputeModule: matches[2],
			NodeBMC:       matches[3],
			NodeBMCNic:    matches[4],
		}
	case xnametypes.NodeEnclosure:
		component = NodeEnclosure{
			Cabinet:       matches[0],
			Chassis:       matches[1],
			ComputeModule: matches[2],
			NodeEnclosure: matches[3],
		}
	case xnametypes.NodeEnclosurePowerSupply:
		component = NodeEnclosurePowerSupply{
			Cabinet:                  matches[0],
			Chassis:                  matches[1],
			ComputeModule:            matches[2],
			NodeEnclosure:            matches[3],
			NodeEnclosurePowerSupply: matches[4],
		}
	case xnametypes.NodeFpga:
		component = NodeFpga{
			Cabinet:       matches[0],
			Chassis:       matches[1],
			ComputeModule: matches[2],
			NodeEnclosure: matches[3],
			NodeFpga:      matches[4],
		}
	case xnametypes.NodePowerConnector:
		component = NodePowerConnector{
			Cabinet:            matches[0],
			Chassis:            matches[1],
			ComputeModule:      matches[2],
			NodePowerConnector: matches[3],
		}
	case xnametypes.MgmtHLSwitchEnclosure:
		component = MgmtHLSwitchEnclosure{
			Cabinet:               matches[0],
			Chassis:               matches[1],
			MgmtHLSwitchEnclosure: matches[2],
		}
	case xnametypes.MgmtHLSwitch:
		component = MgmtHLSwitch{
			Cabinet:               matches[0],
			Chassis:               matches[1],
			MgmtHLSwitchEnclosure: matches[2],
			MgmtHLSwitch:          matches[3],
		}
	case xnametypes.MgmtSwitch:
		component = MgmtSwitch{
			Cabinet:    matches[0],
			Chassis:    matches[1],
			MgmtSwitch: matches[2],
		}
	case xnametypes.MgmtSwitchConnector:
		component = MgmtSwitchConnector{
			Cabinet:             matches[0],
			Chassis:             matches[1],
			MgmtSwitch:          matches[2],
			MgmtSwitchConnector: matches[3],
		}
	case xnametypes.RouterModule:
		component = RouterModule{
			Cabinet:      matches[0],
			Chassis:      matches[1],
			RouterModule: matches[2],
		}
	case xnametypes.HSNAsic:
		component = HSNAsic{
			Cabinet:      matches[0],
			Chassis:      matches[1],
			RouterModule: matches[2],
			HSNAsic:      matches[3],
		}
	case xnametypes.HSNLink:
		component = HSNLink{
			Cabinet:      matches[0],
			Chassis:      matches[1],
			RouterModule: matches[2],
			HSNAsic:      matches[3],
			HSNLink:      matches[4],
		}
	case xnametypes.HSNBoard:
		component = HSNBoard{
			Cabinet:      matches[0],
			Chassis:      matches[1],
			RouterModule: matches[2],
			HSNBoard:     matches[3],
		}
	case xnametypes.HSNConnector:
		component = HSNConnector{
			Cabinet:      matches[0],
			Chassis:      matches[1],
			RouterModule: matches[2],
			HSNConnector: matches[3],
		}
	case xnametypes.HSNConnectorPort:
		component = HSNConnectorPort{
			Cabinet:          matches[0],
			Chassis:          matches[1],
			RouterModule:     matches[2],
			HSNConnector:     matches[3],
			HSNConnectorPort: matches[4],
		}
	case xnametypes.RouterBMC:
		component = RouterBMC{
			Cabinet:      matches[0],
			Chassis:      matches[1],
			RouterModule: matches[2],
			RouterBMC:    matches[3],
		}
	case xnametypes.RouterBMCNic:
		component = RouterBMCNic{
			Cabinet:      matches[0],
			Chassis:      matches[1],
			RouterModule: matches[2],
			RouterBMC:    matches[3],
			RouterBMCNic: matches[4],
		}
	case xnametypes.RouterFpga:
		component = RouterFpga{
			Cabinet:      matches[0],
			Chassis:      matches[1],
			RouterModule: matches[2],
			RouterFpga:   matches[3],
		}
	case xnametypes.RouterPowerConnector:
		component = RouterPowerConnector{
			Cabinet:              matches[0],
			Chassis:              matches[1],
			RouterModule:         matches[2],
			RouterPowerConnector: matches[3],
		}
	case xnametypes.RouterTOR:
		component = RouterTOR{
			Cabinet:      matches[0],
			Chassis:      matches[1],
			RouterModule: matches[2],
			RouterTOR:    matches[3],
		}
	case xnametypes.RouterTORFpga:
		component = RouterTORFpga{
			Cabinet:       matches[0],
			Chassis:       matches[1],
			RouterModule:  matches[2],
			RouterTOR:     matches[3],
			RouterTORFpga: matches[4],
		}
	default:
		return nil
	}
	return component
}
