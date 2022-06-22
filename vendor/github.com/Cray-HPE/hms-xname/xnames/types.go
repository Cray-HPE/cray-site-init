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
	"fmt"

	"github.com/Cray-HPE/hms-xname/xnametypes"
)

// System - sS
type System struct {
}

// Type will return the corresponding HMSType
func (x System) Type() xnametypes.HMSType {
	return xnametypes.System
}

// String will stringify System into the format of sS
func (x System) String() string {
	return fmt.Sprintf(
		"s0",
	)
}

// ParentGeneric will determine the parent of this System, and return it as a GenericXname interface
func (x System) ParentGeneric() GenericXname {

	return nil
}

// CDU will get a child component with the specified ordinal
func (x System) CDU(cDU int) CDU {
	return CDU{
		CDU: cDU,
	}
}

// Cabinet will get a child component with the specified ordinal
func (x System) Cabinet(cabinet int) Cabinet {
	return Cabinet{
		Cabinet: cabinet,
	}
}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x System) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid System xname: %s", xname)
	}

	return nil
}

// IsController returns whether System is a controller type, i.e. that
// would host a Redfish entry point
func (x System) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// CDU - dD
type CDU struct {
	CDU int // dD
}

// Type will return the corresponding HMSType
func (x CDU) Type() xnametypes.HMSType {
	return xnametypes.CDU
}

// String will stringify CDU into the format of dD
func (x CDU) String() string {
	return fmt.Sprintf(
		"d%d",
		x.CDU,
	)
}

// Parent will determine the parent of this CDU
func (x CDU) Parent() System {
	return System{}
}

// ParentGeneric will determine the parent of this CDU, and return it as a GenericXname interface
func (x CDU) ParentGeneric() GenericXname {
	return x.Parent()

}

// CDUMgmtSwitch will get a child component with the specified ordinal
func (x CDU) CDUMgmtSwitch(cDUMgmtSwitch int) CDUMgmtSwitch {
	return CDUMgmtSwitch{
		CDU:           x.CDU,
		CDUMgmtSwitch: cDUMgmtSwitch,
	}
}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x CDU) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid CDU xname: %s", xname)
	}

	return nil
}

// IsController returns whether CDU is a controller type, i.e. that
// would host a Redfish entry point
func (x CDU) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// CDUMgmtSwitch - dDwW
type CDUMgmtSwitch struct {
	CDU           int // dD
	CDUMgmtSwitch int // wW
}

// Type will return the corresponding HMSType
func (x CDUMgmtSwitch) Type() xnametypes.HMSType {
	return xnametypes.CDUMgmtSwitch
}

// String will stringify CDUMgmtSwitch into the format of dDwW
func (x CDUMgmtSwitch) String() string {
	return fmt.Sprintf(
		"d%dw%d",
		x.CDU,
		x.CDUMgmtSwitch,
	)
}

// Parent will determine the parent of this CDUMgmtSwitch
func (x CDUMgmtSwitch) Parent() CDU {
	return CDU{
		CDU: x.CDU,
	}
}

// ParentGeneric will determine the parent of this CDUMgmtSwitch, and return it as a GenericXname interface
func (x CDUMgmtSwitch) ParentGeneric() GenericXname {
	return x.Parent()

}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x CDUMgmtSwitch) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid CDUMgmtSwitch xname: %s", xname)
	}

	return nil
}

// IsController returns whether CDUMgmtSwitch is a controller type, i.e. that
// would host a Redfish entry point
func (x CDUMgmtSwitch) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// Cabinet - xX
type Cabinet struct {
	Cabinet int // xX
}

// Type will return the corresponding HMSType
func (x Cabinet) Type() xnametypes.HMSType {
	return xnametypes.Cabinet
}

// String will stringify Cabinet into the format of xX
func (x Cabinet) String() string {
	return fmt.Sprintf(
		"x%d",
		x.Cabinet,
	)
}

// Parent will determine the parent of this Cabinet
func (x Cabinet) Parent() System {
	return System{}
}

// ParentGeneric will determine the parent of this Cabinet, and return it as a GenericXname interface
func (x Cabinet) ParentGeneric() GenericXname {
	return x.Parent()

}

// CEC will get a child component with the specified ordinal
func (x Cabinet) CEC(cEC int) CEC {
	return CEC{
		Cabinet: x.Cabinet,
		CEC:     cEC,
	}
}

// CabinetBMC will get a child component with the specified ordinal
func (x Cabinet) CabinetBMC(cabinetBMC int) CabinetBMC {
	return CabinetBMC{
		Cabinet:    x.Cabinet,
		CabinetBMC: cabinetBMC,
	}
}

// CabinetCDU will get a child component with the specified ordinal
func (x Cabinet) CabinetCDU(cabinetCDU int) CabinetCDU {
	return CabinetCDU{
		Cabinet:    x.Cabinet,
		CabinetCDU: cabinetCDU,
	}
}

// CabinetPDUController will get a child component with the specified ordinal
func (x Cabinet) CabinetPDUController(cabinetPDUController int) CabinetPDUController {
	return CabinetPDUController{
		Cabinet:              x.Cabinet,
		CabinetPDUController: cabinetPDUController,
	}
}

// Chassis will get a child component with the specified ordinal
func (x Cabinet) Chassis(chassis int) Chassis {
	return Chassis{
		Cabinet: x.Cabinet,
		Chassis: chassis,
	}
}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x Cabinet) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid Cabinet xname: %s", xname)
	}

	return nil
}

// IsController returns whether Cabinet is a controller type, i.e. that
// would host a Redfish entry point
func (x Cabinet) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// CEC - xXeE
type CEC struct {
	Cabinet int // xX
	CEC     int // eE
}

// Type will return the corresponding HMSType
func (x CEC) Type() xnametypes.HMSType {
	return xnametypes.CEC
}

// String will stringify CEC into the format of xXeE
func (x CEC) String() string {
	return fmt.Sprintf(
		"x%de%d",
		x.Cabinet,
		x.CEC,
	)
}

// Parent will determine the parent of this CEC
func (x CEC) Parent() Cabinet {
	return Cabinet{
		Cabinet: x.Cabinet,
	}
}

// ParentGeneric will determine the parent of this CEC, and return it as a GenericXname interface
func (x CEC) ParentGeneric() GenericXname {
	return x.Parent()

}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x CEC) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid CEC xname: %s", xname)
	}

	return nil
}

// IsController returns whether CEC is a controller type, i.e. that
// would host a Redfish entry point
func (x CEC) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// CabinetBMC - xXbB
type CabinetBMC struct {
	Cabinet    int // xX
	CabinetBMC int // bB
}

// Type will return the corresponding HMSType
func (x CabinetBMC) Type() xnametypes.HMSType {
	return xnametypes.CabinetBMC
}

// String will stringify CabinetBMC into the format of xXbB
func (x CabinetBMC) String() string {
	return fmt.Sprintf(
		"x%db%d",
		x.Cabinet,
		x.CabinetBMC,
	)
}

// Parent will determine the parent of this CabinetBMC
func (x CabinetBMC) Parent() Cabinet {
	return Cabinet{
		Cabinet: x.Cabinet,
	}
}

// ParentGeneric will determine the parent of this CabinetBMC, and return it as a GenericXname interface
func (x CabinetBMC) ParentGeneric() GenericXname {
	return x.Parent()

}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x CabinetBMC) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid CabinetBMC xname: %s", xname)
	}

	return nil
}

// IsController returns whether CabinetBMC is a controller type, i.e. that
// would host a Redfish entry point
func (x CabinetBMC) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// CabinetCDU - xXdD
type CabinetCDU struct {
	Cabinet    int // xX
	CabinetCDU int // dD
}

// Type will return the corresponding HMSType
func (x CabinetCDU) Type() xnametypes.HMSType {
	return xnametypes.CabinetCDU
}

// String will stringify CabinetCDU into the format of xXdD
func (x CabinetCDU) String() string {
	return fmt.Sprintf(
		"x%dd%d",
		x.Cabinet,
		x.CabinetCDU,
	)
}

// Parent will determine the parent of this CabinetCDU
func (x CabinetCDU) Parent() Cabinet {
	return Cabinet{
		Cabinet: x.Cabinet,
	}
}

// ParentGeneric will determine the parent of this CabinetCDU, and return it as a GenericXname interface
func (x CabinetCDU) ParentGeneric() GenericXname {
	return x.Parent()

}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x CabinetCDU) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid CabinetCDU xname: %s", xname)
	}

	return nil
}

// IsController returns whether CabinetCDU is a controller type, i.e. that
// would host a Redfish entry point
func (x CabinetCDU) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// CabinetPDUController - xXmM
type CabinetPDUController struct {
	Cabinet              int // xX
	CabinetPDUController int // mM
}

// Type will return the corresponding HMSType
func (x CabinetPDUController) Type() xnametypes.HMSType {
	return xnametypes.CabinetPDUController
}

// String will stringify CabinetPDUController into the format of xXmM
func (x CabinetPDUController) String() string {
	return fmt.Sprintf(
		"x%dm%d",
		x.Cabinet,
		x.CabinetPDUController,
	)
}

// Parent will determine the parent of this CabinetPDUController
func (x CabinetPDUController) Parent() Cabinet {
	return Cabinet{
		Cabinet: x.Cabinet,
	}
}

// ParentGeneric will determine the parent of this CabinetPDUController, and return it as a GenericXname interface
func (x CabinetPDUController) ParentGeneric() GenericXname {
	return x.Parent()

}

// CabinetPDU will get a child component with the specified ordinal
func (x CabinetPDUController) CabinetPDU(cabinetPDU int) CabinetPDU {
	return CabinetPDU{
		Cabinet:              x.Cabinet,
		CabinetPDUController: x.CabinetPDUController,
		CabinetPDU:           cabinetPDU,
	}
}

// CabinetPDUNic will get a child component with the specified ordinal
func (x CabinetPDUController) CabinetPDUNic(cabinetPDUNic int) CabinetPDUNic {
	return CabinetPDUNic{
		Cabinet:              x.Cabinet,
		CabinetPDUController: x.CabinetPDUController,
		CabinetPDUNic:        cabinetPDUNic,
	}
}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x CabinetPDUController) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid CabinetPDUController xname: %s", xname)
	}

	return nil
}

// IsController returns whether CabinetPDUController is a controller type, i.e. that
// would host a Redfish entry point
func (x CabinetPDUController) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// CabinetPDU - xXmMpP
type CabinetPDU struct {
	Cabinet              int // xX
	CabinetPDUController int // mM
	CabinetPDU           int // pP
}

// Type will return the corresponding HMSType
func (x CabinetPDU) Type() xnametypes.HMSType {
	return xnametypes.CabinetPDU
}

// String will stringify CabinetPDU into the format of xXmMpP
func (x CabinetPDU) String() string {
	return fmt.Sprintf(
		"x%dm%dp%d",
		x.Cabinet,
		x.CabinetPDUController,
		x.CabinetPDU,
	)
}

// Parent will determine the parent of this CabinetPDU
func (x CabinetPDU) Parent() CabinetPDUController {
	return CabinetPDUController{
		Cabinet:              x.Cabinet,
		CabinetPDUController: x.CabinetPDUController,
	}
}

// ParentGeneric will determine the parent of this CabinetPDU, and return it as a GenericXname interface
func (x CabinetPDU) ParentGeneric() GenericXname {
	return x.Parent()

}

// CabinetPDUOutlet will get a child component with the specified ordinal
func (x CabinetPDU) CabinetPDUOutlet(cabinetPDUOutlet int) CabinetPDUOutlet {
	return CabinetPDUOutlet{
		Cabinet:              x.Cabinet,
		CabinetPDUController: x.CabinetPDUController,
		CabinetPDU:           x.CabinetPDU,
		CabinetPDUOutlet:     cabinetPDUOutlet,
	}
}

// CabinetPDUPowerConnector will get a child component with the specified ordinal
func (x CabinetPDU) CabinetPDUPowerConnector(cabinetPDUPowerConnector int) CabinetPDUPowerConnector {
	return CabinetPDUPowerConnector{
		Cabinet:                  x.Cabinet,
		CabinetPDUController:     x.CabinetPDUController,
		CabinetPDU:               x.CabinetPDU,
		CabinetPDUPowerConnector: cabinetPDUPowerConnector,
	}
}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x CabinetPDU) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid CabinetPDU xname: %s", xname)
	}

	return nil
}

// IsController returns whether CabinetPDU is a controller type, i.e. that
// would host a Redfish entry point
func (x CabinetPDU) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// CabinetPDUOutlet - xXmMpPjJ
type CabinetPDUOutlet struct {
	Cabinet              int // xX
	CabinetPDUController int // mM
	CabinetPDU           int // pP
	CabinetPDUOutlet     int // jJ
}

// Type will return the corresponding HMSType
func (x CabinetPDUOutlet) Type() xnametypes.HMSType {
	return xnametypes.CabinetPDUOutlet
}

// String will stringify CabinetPDUOutlet into the format of xXmMpPjJ
func (x CabinetPDUOutlet) String() string {
	return fmt.Sprintf(
		"x%dm%dp%dj%d",
		x.Cabinet,
		x.CabinetPDUController,
		x.CabinetPDU,
		x.CabinetPDUOutlet,
	)
}

// Parent will determine the parent of this CabinetPDUOutlet
func (x CabinetPDUOutlet) Parent() CabinetPDU {
	return CabinetPDU{
		Cabinet:              x.Cabinet,
		CabinetPDUController: x.CabinetPDUController,
		CabinetPDU:           x.CabinetPDU,
	}
}

// ParentGeneric will determine the parent of this CabinetPDUOutlet, and return it as a GenericXname interface
func (x CabinetPDUOutlet) ParentGeneric() GenericXname {
	return x.Parent()

}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x CabinetPDUOutlet) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid CabinetPDUOutlet xname: %s", xname)
	}

	return nil
}

// IsController returns whether CabinetPDUOutlet is a controller type, i.e. that
// would host a Redfish entry point
func (x CabinetPDUOutlet) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// CabinetPDUPowerConnector - xXmMpPvV
type CabinetPDUPowerConnector struct {
	Cabinet                  int // xX
	CabinetPDUController     int // mM
	CabinetPDU               int // pP
	CabinetPDUPowerConnector int // vV
}

// Type will return the corresponding HMSType
func (x CabinetPDUPowerConnector) Type() xnametypes.HMSType {
	return xnametypes.CabinetPDUPowerConnector
}

// String will stringify CabinetPDUPowerConnector into the format of xXmMpPvV
func (x CabinetPDUPowerConnector) String() string {
	return fmt.Sprintf(
		"x%dm%dp%dv%d",
		x.Cabinet,
		x.CabinetPDUController,
		x.CabinetPDU,
		x.CabinetPDUPowerConnector,
	)
}

// Parent will determine the parent of this CabinetPDUPowerConnector
func (x CabinetPDUPowerConnector) Parent() CabinetPDU {
	return CabinetPDU{
		Cabinet:              x.Cabinet,
		CabinetPDUController: x.CabinetPDUController,
		CabinetPDU:           x.CabinetPDU,
	}
}

// ParentGeneric will determine the parent of this CabinetPDUPowerConnector, and return it as a GenericXname interface
func (x CabinetPDUPowerConnector) ParentGeneric() GenericXname {
	return x.Parent()

}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x CabinetPDUPowerConnector) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid CabinetPDUPowerConnector xname: %s", xname)
	}

	return nil
}

// IsController returns whether CabinetPDUPowerConnector is a controller type, i.e. that
// would host a Redfish entry point
func (x CabinetPDUPowerConnector) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// CabinetPDUNic - xXmMpPiI
type CabinetPDUNic struct {
	Cabinet              int // xX
	CabinetPDUController int // mM
	CabinetPDUNic        int // iI
}

// Type will return the corresponding HMSType
func (x CabinetPDUNic) Type() xnametypes.HMSType {
	return xnametypes.CabinetPDUNic
}

// String will stringify CabinetPDUNic into the format of xXmMpPiI
func (x CabinetPDUNic) String() string {
	return fmt.Sprintf(
		"x%dm%di%d",
		x.Cabinet,
		x.CabinetPDUController,
		x.CabinetPDUNic,
	)
}

// Parent will determine the parent of this CabinetPDUNic
func (x CabinetPDUNic) Parent() CabinetPDUController {
	return CabinetPDUController{
		Cabinet:              x.Cabinet,
		CabinetPDUController: x.CabinetPDUController,
	}
}

// ParentGeneric will determine the parent of this CabinetPDUNic, and return it as a GenericXname interface
func (x CabinetPDUNic) ParentGeneric() GenericXname {
	return x.Parent()

}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x CabinetPDUNic) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid CabinetPDUNic xname: %s", xname)
	}

	return nil
}

// IsController returns whether CabinetPDUNic is a controller type, i.e. that
// would host a Redfish entry point
func (x CabinetPDUNic) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// Chassis - xXcC
type Chassis struct {
	Cabinet int // xX
	Chassis int // cC
}

// Type will return the corresponding HMSType
func (x Chassis) Type() xnametypes.HMSType {
	return xnametypes.Chassis
}

// String will stringify Chassis into the format of xXcC
func (x Chassis) String() string {
	return fmt.Sprintf(
		"x%dc%d",
		x.Cabinet,
		x.Chassis,
	)
}

// Parent will determine the parent of this Chassis
func (x Chassis) Parent() Cabinet {
	return Cabinet{
		Cabinet: x.Cabinet,
	}
}

// ParentGeneric will determine the parent of this Chassis, and return it as a GenericXname interface
func (x Chassis) ParentGeneric() GenericXname {
	return x.Parent()

}

// CMMFpga will get a child component with the specified ordinal
func (x Chassis) CMMFpga(cMMFpga int) CMMFpga {
	return CMMFpga{
		Cabinet: x.Cabinet,
		Chassis: x.Chassis,
		CMMFpga: cMMFpga,
	}
}

// CMMRectifier will get a child component with the specified ordinal
func (x Chassis) CMMRectifier(cMMRectifier int) CMMRectifier {
	return CMMRectifier{
		Cabinet:      x.Cabinet,
		Chassis:      x.Chassis,
		CMMRectifier: cMMRectifier,
	}
}

// ChassisBMC will get a child component with the specified ordinal
func (x Chassis) ChassisBMC(chassisBMC int) ChassisBMC {
	return ChassisBMC{
		Cabinet:    x.Cabinet,
		Chassis:    x.Chassis,
		ChassisBMC: chassisBMC,
	}
}

// ComputeModule will get a child component with the specified ordinal
func (x Chassis) ComputeModule(computeModule int) ComputeModule {
	return ComputeModule{
		Cabinet:       x.Cabinet,
		Chassis:       x.Chassis,
		ComputeModule: computeModule,
	}
}

// MgmtHLSwitchEnclosure will get a child component with the specified ordinal
func (x Chassis) MgmtHLSwitchEnclosure(mgmtHLSwitchEnclosure int) MgmtHLSwitchEnclosure {
	return MgmtHLSwitchEnclosure{
		Cabinet:               x.Cabinet,
		Chassis:               x.Chassis,
		MgmtHLSwitchEnclosure: mgmtHLSwitchEnclosure,
	}
}

// MgmtSwitch will get a child component with the specified ordinal
func (x Chassis) MgmtSwitch(mgmtSwitch int) MgmtSwitch {
	return MgmtSwitch{
		Cabinet:    x.Cabinet,
		Chassis:    x.Chassis,
		MgmtSwitch: mgmtSwitch,
	}
}

// RouterModule will get a child component with the specified ordinal
func (x Chassis) RouterModule(routerModule int) RouterModule {
	return RouterModule{
		Cabinet:      x.Cabinet,
		Chassis:      x.Chassis,
		RouterModule: routerModule,
	}
}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x Chassis) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid Chassis xname: %s", xname)
	}

	return nil
}

// IsController returns whether Chassis is a controller type, i.e. that
// would host a Redfish entry point
func (x Chassis) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// CMMFpga - xXcCfF
type CMMFpga struct {
	Cabinet int // xX
	Chassis int // cC
	CMMFpga int // fF
}

// Type will return the corresponding HMSType
func (x CMMFpga) Type() xnametypes.HMSType {
	return xnametypes.CMMFpga
}

// String will stringify CMMFpga into the format of xXcCfF
func (x CMMFpga) String() string {
	return fmt.Sprintf(
		"x%dc%df%d",
		x.Cabinet,
		x.Chassis,
		x.CMMFpga,
	)
}

// Parent will determine the parent of this CMMFpga
func (x CMMFpga) Parent() Chassis {
	return Chassis{
		Cabinet: x.Cabinet,
		Chassis: x.Chassis,
	}
}

// ParentGeneric will determine the parent of this CMMFpga, and return it as a GenericXname interface
func (x CMMFpga) ParentGeneric() GenericXname {
	return x.Parent()

}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x CMMFpga) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid CMMFpga xname: %s", xname)
	}

	return nil
}

// IsController returns whether CMMFpga is a controller type, i.e. that
// would host a Redfish entry point
func (x CMMFpga) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// CMMRectifier - xXcCtT
type CMMRectifier struct {
	Cabinet      int // xX
	Chassis      int // cC
	CMMRectifier int // tT
}

// Type will return the corresponding HMSType
func (x CMMRectifier) Type() xnametypes.HMSType {
	return xnametypes.CMMRectifier
}

// String will stringify CMMRectifier into the format of xXcCtT
func (x CMMRectifier) String() string {
	return fmt.Sprintf(
		"x%dc%dt%d",
		x.Cabinet,
		x.Chassis,
		x.CMMRectifier,
	)
}

// Parent will determine the parent of this CMMRectifier
func (x CMMRectifier) Parent() Chassis {
	return Chassis{
		Cabinet: x.Cabinet,
		Chassis: x.Chassis,
	}
}

// ParentGeneric will determine the parent of this CMMRectifier, and return it as a GenericXname interface
func (x CMMRectifier) ParentGeneric() GenericXname {
	return x.Parent()

}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x CMMRectifier) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid CMMRectifier xname: %s", xname)
	}

	return nil
}

// IsController returns whether CMMRectifier is a controller type, i.e. that
// would host a Redfish entry point
func (x CMMRectifier) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// ChassisBMC - xXcCbB
type ChassisBMC struct {
	Cabinet    int // xX
	Chassis    int // cC
	ChassisBMC int // bB
}

// Type will return the corresponding HMSType
func (x ChassisBMC) Type() xnametypes.HMSType {
	return xnametypes.ChassisBMC
}

// String will stringify ChassisBMC into the format of xXcCbB
func (x ChassisBMC) String() string {
	return fmt.Sprintf(
		"x%dc%db%d",
		x.Cabinet,
		x.Chassis,
		x.ChassisBMC,
	)
}

// Parent will determine the parent of this ChassisBMC
func (x ChassisBMC) Parent() Chassis {
	return Chassis{
		Cabinet: x.Cabinet,
		Chassis: x.Chassis,
	}
}

// ParentGeneric will determine the parent of this ChassisBMC, and return it as a GenericXname interface
func (x ChassisBMC) ParentGeneric() GenericXname {
	return x.Parent()

}

// ChassisBMCNic will get a child component with the specified ordinal
func (x ChassisBMC) ChassisBMCNic(chassisBMCNic int) ChassisBMCNic {
	return ChassisBMCNic{
		Cabinet:       x.Cabinet,
		Chassis:       x.Chassis,
		ChassisBMC:    x.ChassisBMC,
		ChassisBMCNic: chassisBMCNic,
	}
}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x ChassisBMC) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid ChassisBMC xname: %s", xname)
	}

	return nil
}

// IsController returns whether ChassisBMC is a controller type, i.e. that
// would host a Redfish entry point
func (x ChassisBMC) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// ChassisBMCNic - xXcCbBiI
type ChassisBMCNic struct {
	Cabinet       int // xX
	Chassis       int // cC
	ChassisBMC    int // bB
	ChassisBMCNic int // iI
}

// Type will return the corresponding HMSType
func (x ChassisBMCNic) Type() xnametypes.HMSType {
	return xnametypes.ChassisBMCNic
}

// String will stringify ChassisBMCNic into the format of xXcCbBiI
func (x ChassisBMCNic) String() string {
	return fmt.Sprintf(
		"x%dc%db%di%d",
		x.Cabinet,
		x.Chassis,
		x.ChassisBMC,
		x.ChassisBMCNic,
	)
}

// Parent will determine the parent of this ChassisBMCNic
func (x ChassisBMCNic) Parent() ChassisBMC {
	return ChassisBMC{
		Cabinet:    x.Cabinet,
		Chassis:    x.Chassis,
		ChassisBMC: x.ChassisBMC,
	}
}

// ParentGeneric will determine the parent of this ChassisBMCNic, and return it as a GenericXname interface
func (x ChassisBMCNic) ParentGeneric() GenericXname {
	return x.Parent()

}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x ChassisBMCNic) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid ChassisBMCNic xname: %s", xname)
	}

	return nil
}

// IsController returns whether ChassisBMCNic is a controller type, i.e. that
// would host a Redfish entry point
func (x ChassisBMCNic) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// ComputeModule - xXcCsS
type ComputeModule struct {
	Cabinet       int // xX
	Chassis       int // cC
	ComputeModule int // sS
}

// Type will return the corresponding HMSType
func (x ComputeModule) Type() xnametypes.HMSType {
	return xnametypes.ComputeModule
}

// String will stringify ComputeModule into the format of xXcCsS
func (x ComputeModule) String() string {
	return fmt.Sprintf(
		"x%dc%ds%d",
		x.Cabinet,
		x.Chassis,
		x.ComputeModule,
	)
}

// Parent will determine the parent of this ComputeModule
func (x ComputeModule) Parent() Chassis {
	return Chassis{
		Cabinet: x.Cabinet,
		Chassis: x.Chassis,
	}
}

// ParentGeneric will determine the parent of this ComputeModule, and return it as a GenericXname interface
func (x ComputeModule) ParentGeneric() GenericXname {
	return x.Parent()

}

// NodeBMC will get a child component with the specified ordinal
func (x ComputeModule) NodeBMC(nodeBMC int) NodeBMC {
	return NodeBMC{
		Cabinet:       x.Cabinet,
		Chassis:       x.Chassis,
		ComputeModule: x.ComputeModule,
		NodeBMC:       nodeBMC,
	}
}

// NodeEnclosure will get a child component with the specified ordinal
func (x ComputeModule) NodeEnclosure(nodeEnclosure int) NodeEnclosure {
	return NodeEnclosure{
		Cabinet:       x.Cabinet,
		Chassis:       x.Chassis,
		ComputeModule: x.ComputeModule,
		NodeEnclosure: nodeEnclosure,
	}
}

// NodePowerConnector will get a child component with the specified ordinal
func (x ComputeModule) NodePowerConnector(nodePowerConnector int) NodePowerConnector {
	return NodePowerConnector{
		Cabinet:            x.Cabinet,
		Chassis:            x.Chassis,
		ComputeModule:      x.ComputeModule,
		NodePowerConnector: nodePowerConnector,
	}
}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x ComputeModule) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid ComputeModule xname: %s", xname)
	}

	return nil
}

// IsController returns whether ComputeModule is a controller type, i.e. that
// would host a Redfish entry point
func (x ComputeModule) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// NodeBMC - xXcCsSbB
type NodeBMC struct {
	Cabinet       int // xX
	Chassis       int // cC
	ComputeModule int // sS
	NodeBMC       int // bB
}

// Type will return the corresponding HMSType
func (x NodeBMC) Type() xnametypes.HMSType {
	return xnametypes.NodeBMC
}

// String will stringify NodeBMC into the format of xXcCsSbB
func (x NodeBMC) String() string {
	return fmt.Sprintf(
		"x%dc%ds%db%d",
		x.Cabinet,
		x.Chassis,
		x.ComputeModule,
		x.NodeBMC,
	)
}

// Parent will determine the parent of this NodeBMC
func (x NodeBMC) Parent() ComputeModule {
	return ComputeModule{
		Cabinet:       x.Cabinet,
		Chassis:       x.Chassis,
		ComputeModule: x.ComputeModule,
	}
}

// ParentGeneric will determine the parent of this NodeBMC, and return it as a GenericXname interface
func (x NodeBMC) ParentGeneric() GenericXname {
	return x.Parent()

}

// Node will get a child component with the specified ordinal
func (x NodeBMC) Node(node int) Node {
	return Node{
		Cabinet:       x.Cabinet,
		Chassis:       x.Chassis,
		ComputeModule: x.ComputeModule,
		NodeBMC:       x.NodeBMC,
		Node:          node,
	}
}

// NodeBMCNic will get a child component with the specified ordinal
func (x NodeBMC) NodeBMCNic(nodeBMCNic int) NodeBMCNic {
	return NodeBMCNic{
		Cabinet:       x.Cabinet,
		Chassis:       x.Chassis,
		ComputeModule: x.ComputeModule,
		NodeBMC:       x.NodeBMC,
		NodeBMCNic:    nodeBMCNic,
	}
}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x NodeBMC) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid NodeBMC xname: %s", xname)
	}

	return nil
}

// IsController returns whether NodeBMC is a controller type, i.e. that
// would host a Redfish entry point
func (x NodeBMC) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// Node - xXcCsSbBnN
type Node struct {
	Cabinet       int // xX
	Chassis       int // cC
	ComputeModule int // sS
	NodeBMC       int // bB
	Node          int // nN
}

// Type will return the corresponding HMSType
func (x Node) Type() xnametypes.HMSType {
	return xnametypes.Node
}

// String will stringify Node into the format of xXcCsSbBnN
func (x Node) String() string {
	return fmt.Sprintf(
		"x%dc%ds%db%dn%d",
		x.Cabinet,
		x.Chassis,
		x.ComputeModule,
		x.NodeBMC,
		x.Node,
	)
}

// Parent will determine the parent of this Node
func (x Node) Parent() NodeBMC {
	return NodeBMC{
		Cabinet:       x.Cabinet,
		Chassis:       x.Chassis,
		ComputeModule: x.ComputeModule,
		NodeBMC:       x.NodeBMC,
	}
}

// ParentGeneric will determine the parent of this Node, and return it as a GenericXname interface
func (x Node) ParentGeneric() GenericXname {
	return x.Parent()

}

// Memory will get a child component with the specified ordinal
func (x Node) Memory(memory int) Memory {
	return Memory{
		Cabinet:       x.Cabinet,
		Chassis:       x.Chassis,
		ComputeModule: x.ComputeModule,
		NodeBMC:       x.NodeBMC,
		Node:          x.Node,
		Memory:        memory,
	}
}

// NodeAccel will get a child component with the specified ordinal
func (x Node) NodeAccel(nodeAccel int) NodeAccel {
	return NodeAccel{
		Cabinet:       x.Cabinet,
		Chassis:       x.Chassis,
		ComputeModule: x.ComputeModule,
		NodeBMC:       x.NodeBMC,
		Node:          x.Node,
		NodeAccel:     nodeAccel,
	}
}

// NodeAccelRiser will get a child component with the specified ordinal
func (x Node) NodeAccelRiser(nodeAccelRiser int) NodeAccelRiser {
	return NodeAccelRiser{
		Cabinet:        x.Cabinet,
		Chassis:        x.Chassis,
		ComputeModule:  x.ComputeModule,
		NodeBMC:        x.NodeBMC,
		Node:           x.Node,
		NodeAccelRiser: nodeAccelRiser,
	}
}

// NodeHsnNic will get a child component with the specified ordinal
func (x Node) NodeHsnNic(nodeHsnNic int) NodeHsnNic {
	return NodeHsnNic{
		Cabinet:       x.Cabinet,
		Chassis:       x.Chassis,
		ComputeModule: x.ComputeModule,
		NodeBMC:       x.NodeBMC,
		Node:          x.Node,
		NodeHsnNic:    nodeHsnNic,
	}
}

// NodeNic will get a child component with the specified ordinal
func (x Node) NodeNic(nodeNic int) NodeNic {
	return NodeNic{
		Cabinet:       x.Cabinet,
		Chassis:       x.Chassis,
		ComputeModule: x.ComputeModule,
		NodeBMC:       x.NodeBMC,
		Node:          x.Node,
		NodeNic:       nodeNic,
	}
}

// Processor will get a child component with the specified ordinal
func (x Node) Processor(processor int) Processor {
	return Processor{
		Cabinet:       x.Cabinet,
		Chassis:       x.Chassis,
		ComputeModule: x.ComputeModule,
		NodeBMC:       x.NodeBMC,
		Node:          x.Node,
		Processor:     processor,
	}
}

// StorageGroup will get a child component with the specified ordinal
func (x Node) StorageGroup(storageGroup int) StorageGroup {
	return StorageGroup{
		Cabinet:       x.Cabinet,
		Chassis:       x.Chassis,
		ComputeModule: x.ComputeModule,
		NodeBMC:       x.NodeBMC,
		Node:          x.Node,
		StorageGroup:  storageGroup,
	}
}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x Node) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid Node xname: %s", xname)
	}

	return nil
}

// IsController returns whether Node is a controller type, i.e. that
// would host a Redfish entry point
func (x Node) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// Memory - xXcCsSbBnNdD
type Memory struct {
	Cabinet       int // xX
	Chassis       int // cC
	ComputeModule int // sS
	NodeBMC       int // bB
	Node          int // nN
	Memory        int // dD
}

// Type will return the corresponding HMSType
func (x Memory) Type() xnametypes.HMSType {
	return xnametypes.Memory
}

// String will stringify Memory into the format of xXcCsSbBnNdD
func (x Memory) String() string {
	return fmt.Sprintf(
		"x%dc%ds%db%dn%dd%d",
		x.Cabinet,
		x.Chassis,
		x.ComputeModule,
		x.NodeBMC,
		x.Node,
		x.Memory,
	)
}

// Parent will determine the parent of this Memory
func (x Memory) Parent() Node {
	return Node{
		Cabinet:       x.Cabinet,
		Chassis:       x.Chassis,
		ComputeModule: x.ComputeModule,
		NodeBMC:       x.NodeBMC,
		Node:          x.Node,
	}
}

// ParentGeneric will determine the parent of this Memory, and return it as a GenericXname interface
func (x Memory) ParentGeneric() GenericXname {
	return x.Parent()

}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x Memory) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid Memory xname: %s", xname)
	}

	return nil
}

// IsController returns whether Memory is a controller type, i.e. that
// would host a Redfish entry point
func (x Memory) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// NodeAccel - xXcCsSbBnNaA
type NodeAccel struct {
	Cabinet       int // xX
	Chassis       int // cC
	ComputeModule int // sS
	NodeBMC       int // bB
	Node          int // nN
	NodeAccel     int // aA
}

// Type will return the corresponding HMSType
func (x NodeAccel) Type() xnametypes.HMSType {
	return xnametypes.NodeAccel
}

// String will stringify NodeAccel into the format of xXcCsSbBnNaA
func (x NodeAccel) String() string {
	return fmt.Sprintf(
		"x%dc%ds%db%dn%da%d",
		x.Cabinet,
		x.Chassis,
		x.ComputeModule,
		x.NodeBMC,
		x.Node,
		x.NodeAccel,
	)
}

// Parent will determine the parent of this NodeAccel
func (x NodeAccel) Parent() Node {
	return Node{
		Cabinet:       x.Cabinet,
		Chassis:       x.Chassis,
		ComputeModule: x.ComputeModule,
		NodeBMC:       x.NodeBMC,
		Node:          x.Node,
	}
}

// ParentGeneric will determine the parent of this NodeAccel, and return it as a GenericXname interface
func (x NodeAccel) ParentGeneric() GenericXname {
	return x.Parent()

}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x NodeAccel) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid NodeAccel xname: %s", xname)
	}

	return nil
}

// IsController returns whether NodeAccel is a controller type, i.e. that
// would host a Redfish entry point
func (x NodeAccel) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// NodeAccelRiser - xXcCsSbBnNrR
type NodeAccelRiser struct {
	Cabinet        int // xX
	Chassis        int // cC
	ComputeModule  int // sS
	NodeBMC        int // bB
	Node           int // nN
	NodeAccelRiser int // rR
}

// Type will return the corresponding HMSType
func (x NodeAccelRiser) Type() xnametypes.HMSType {
	return xnametypes.NodeAccelRiser
}

// String will stringify NodeAccelRiser into the format of xXcCsSbBnNrR
func (x NodeAccelRiser) String() string {
	return fmt.Sprintf(
		"x%dc%ds%db%dn%dr%d",
		x.Cabinet,
		x.Chassis,
		x.ComputeModule,
		x.NodeBMC,
		x.Node,
		x.NodeAccelRiser,
	)
}

// Parent will determine the parent of this NodeAccelRiser
func (x NodeAccelRiser) Parent() Node {
	return Node{
		Cabinet:       x.Cabinet,
		Chassis:       x.Chassis,
		ComputeModule: x.ComputeModule,
		NodeBMC:       x.NodeBMC,
		Node:          x.Node,
	}
}

// ParentGeneric will determine the parent of this NodeAccelRiser, and return it as a GenericXname interface
func (x NodeAccelRiser) ParentGeneric() GenericXname {
	return x.Parent()

}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x NodeAccelRiser) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid NodeAccelRiser xname: %s", xname)
	}

	return nil
}

// IsController returns whether NodeAccelRiser is a controller type, i.e. that
// would host a Redfish entry point
func (x NodeAccelRiser) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// NodeHsnNic - xXcCsSbBnNhH
type NodeHsnNic struct {
	Cabinet       int // xX
	Chassis       int // cC
	ComputeModule int // sS
	NodeBMC       int // bB
	Node          int // nN
	NodeHsnNic    int // hH
}

// Type will return the corresponding HMSType
func (x NodeHsnNic) Type() xnametypes.HMSType {
	return xnametypes.NodeHsnNic
}

// String will stringify NodeHsnNic into the format of xXcCsSbBnNhH
func (x NodeHsnNic) String() string {
	return fmt.Sprintf(
		"x%dc%ds%db%dn%dh%d",
		x.Cabinet,
		x.Chassis,
		x.ComputeModule,
		x.NodeBMC,
		x.Node,
		x.NodeHsnNic,
	)
}

// Parent will determine the parent of this NodeHsnNic
func (x NodeHsnNic) Parent() Node {
	return Node{
		Cabinet:       x.Cabinet,
		Chassis:       x.Chassis,
		ComputeModule: x.ComputeModule,
		NodeBMC:       x.NodeBMC,
		Node:          x.Node,
	}
}

// ParentGeneric will determine the parent of this NodeHsnNic, and return it as a GenericXname interface
func (x NodeHsnNic) ParentGeneric() GenericXname {
	return x.Parent()

}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x NodeHsnNic) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid NodeHsnNic xname: %s", xname)
	}

	return nil
}

// IsController returns whether NodeHsnNic is a controller type, i.e. that
// would host a Redfish entry point
func (x NodeHsnNic) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// NodeNic - xXcCsSbBnNiI
type NodeNic struct {
	Cabinet       int // xX
	Chassis       int // cC
	ComputeModule int // sS
	NodeBMC       int // bB
	Node          int // nN
	NodeNic       int // iI
}

// Type will return the corresponding HMSType
func (x NodeNic) Type() xnametypes.HMSType {
	return xnametypes.NodeNic
}

// String will stringify NodeNic into the format of xXcCsSbBnNiI
func (x NodeNic) String() string {
	return fmt.Sprintf(
		"x%dc%ds%db%dn%di%d",
		x.Cabinet,
		x.Chassis,
		x.ComputeModule,
		x.NodeBMC,
		x.Node,
		x.NodeNic,
	)
}

// Parent will determine the parent of this NodeNic
func (x NodeNic) Parent() Node {
	return Node{
		Cabinet:       x.Cabinet,
		Chassis:       x.Chassis,
		ComputeModule: x.ComputeModule,
		NodeBMC:       x.NodeBMC,
		Node:          x.Node,
	}
}

// ParentGeneric will determine the parent of this NodeNic, and return it as a GenericXname interface
func (x NodeNic) ParentGeneric() GenericXname {
	return x.Parent()

}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x NodeNic) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid NodeNic xname: %s", xname)
	}

	return nil
}

// IsController returns whether NodeNic is a controller type, i.e. that
// would host a Redfish entry point
func (x NodeNic) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// Processor - xXcCsSbBnNpP
type Processor struct {
	Cabinet       int // xX
	Chassis       int // cC
	ComputeModule int // sS
	NodeBMC       int // bB
	Node          int // nN
	Processor     int // pP
}

// Type will return the corresponding HMSType
func (x Processor) Type() xnametypes.HMSType {
	return xnametypes.Processor
}

// String will stringify Processor into the format of xXcCsSbBnNpP
func (x Processor) String() string {
	return fmt.Sprintf(
		"x%dc%ds%db%dn%dp%d",
		x.Cabinet,
		x.Chassis,
		x.ComputeModule,
		x.NodeBMC,
		x.Node,
		x.Processor,
	)
}

// Parent will determine the parent of this Processor
func (x Processor) Parent() Node {
	return Node{
		Cabinet:       x.Cabinet,
		Chassis:       x.Chassis,
		ComputeModule: x.ComputeModule,
		NodeBMC:       x.NodeBMC,
		Node:          x.Node,
	}
}

// ParentGeneric will determine the parent of this Processor, and return it as a GenericXname interface
func (x Processor) ParentGeneric() GenericXname {
	return x.Parent()

}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x Processor) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid Processor xname: %s", xname)
	}

	return nil
}

// IsController returns whether Processor is a controller type, i.e. that
// would host a Redfish entry point
func (x Processor) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// StorageGroup - xXcCsSbBnNgG
type StorageGroup struct {
	Cabinet       int // xX
	Chassis       int // cC
	ComputeModule int // sS
	NodeBMC       int // bB
	Node          int // nN
	StorageGroup  int // gG
}

// Type will return the corresponding HMSType
func (x StorageGroup) Type() xnametypes.HMSType {
	return xnametypes.StorageGroup
}

// String will stringify StorageGroup into the format of xXcCsSbBnNgG
func (x StorageGroup) String() string {
	return fmt.Sprintf(
		"x%dc%ds%db%dn%dg%d",
		x.Cabinet,
		x.Chassis,
		x.ComputeModule,
		x.NodeBMC,
		x.Node,
		x.StorageGroup,
	)
}

// Parent will determine the parent of this StorageGroup
func (x StorageGroup) Parent() Node {
	return Node{
		Cabinet:       x.Cabinet,
		Chassis:       x.Chassis,
		ComputeModule: x.ComputeModule,
		NodeBMC:       x.NodeBMC,
		Node:          x.Node,
	}
}

// ParentGeneric will determine the parent of this StorageGroup, and return it as a GenericXname interface
func (x StorageGroup) ParentGeneric() GenericXname {
	return x.Parent()

}

// Drive will get a child component with the specified ordinal
func (x StorageGroup) Drive(drive int) Drive {
	return Drive{
		Cabinet:       x.Cabinet,
		Chassis:       x.Chassis,
		ComputeModule: x.ComputeModule,
		NodeBMC:       x.NodeBMC,
		Node:          x.Node,
		StorageGroup:  x.StorageGroup,
		Drive:         drive,
	}
}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x StorageGroup) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid StorageGroup xname: %s", xname)
	}

	return nil
}

// IsController returns whether StorageGroup is a controller type, i.e. that
// would host a Redfish entry point
func (x StorageGroup) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// Drive - xXcCsSbBnNgGkK
type Drive struct {
	Cabinet       int // xX
	Chassis       int // cC
	ComputeModule int // sS
	NodeBMC       int // bB
	Node          int // nN
	StorageGroup  int // gG
	Drive         int // kK
}

// Type will return the corresponding HMSType
func (x Drive) Type() xnametypes.HMSType {
	return xnametypes.Drive
}

// String will stringify Drive into the format of xXcCsSbBnNgGkK
func (x Drive) String() string {
	return fmt.Sprintf(
		"x%dc%ds%db%dn%dg%dk%d",
		x.Cabinet,
		x.Chassis,
		x.ComputeModule,
		x.NodeBMC,
		x.Node,
		x.StorageGroup,
		x.Drive,
	)
}

// Parent will determine the parent of this Drive
func (x Drive) Parent() StorageGroup {
	return StorageGroup{
		Cabinet:       x.Cabinet,
		Chassis:       x.Chassis,
		ComputeModule: x.ComputeModule,
		NodeBMC:       x.NodeBMC,
		Node:          x.Node,
		StorageGroup:  x.StorageGroup,
	}
}

// ParentGeneric will determine the parent of this Drive, and return it as a GenericXname interface
func (x Drive) ParentGeneric() GenericXname {
	return x.Parent()

}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x Drive) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid Drive xname: %s", xname)
	}

	return nil
}

// IsController returns whether Drive is a controller type, i.e. that
// would host a Redfish entry point
func (x Drive) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// NodeBMCNic - xXcCsSbBiI
type NodeBMCNic struct {
	Cabinet       int // xX
	Chassis       int // cC
	ComputeModule int // sS
	NodeBMC       int // bB
	NodeBMCNic    int // iI
}

// Type will return the corresponding HMSType
func (x NodeBMCNic) Type() xnametypes.HMSType {
	return xnametypes.NodeBMCNic
}

// String will stringify NodeBMCNic into the format of xXcCsSbBiI
func (x NodeBMCNic) String() string {
	return fmt.Sprintf(
		"x%dc%ds%db%di%d",
		x.Cabinet,
		x.Chassis,
		x.ComputeModule,
		x.NodeBMC,
		x.NodeBMCNic,
	)
}

// Parent will determine the parent of this NodeBMCNic
func (x NodeBMCNic) Parent() NodeBMC {
	return NodeBMC{
		Cabinet:       x.Cabinet,
		Chassis:       x.Chassis,
		ComputeModule: x.ComputeModule,
		NodeBMC:       x.NodeBMC,
	}
}

// ParentGeneric will determine the parent of this NodeBMCNic, and return it as a GenericXname interface
func (x NodeBMCNic) ParentGeneric() GenericXname {
	return x.Parent()

}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x NodeBMCNic) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid NodeBMCNic xname: %s", xname)
	}

	return nil
}

// IsController returns whether NodeBMCNic is a controller type, i.e. that
// would host a Redfish entry point
func (x NodeBMCNic) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// NodeEnclosure - xXcCsSbBeE
type NodeEnclosure struct {
	Cabinet       int // xX
	Chassis       int // cC
	ComputeModule int // sS
	NodeEnclosure int // eE
}

// Type will return the corresponding HMSType
func (x NodeEnclosure) Type() xnametypes.HMSType {
	return xnametypes.NodeEnclosure
}

// String will stringify NodeEnclosure into the format of xXcCsSbBeE
func (x NodeEnclosure) String() string {
	return fmt.Sprintf(
		"x%dc%ds%de%d",
		x.Cabinet,
		x.Chassis,
		x.ComputeModule,
		x.NodeEnclosure,
	)
}

// Parent will determine the parent of this NodeEnclosure
func (x NodeEnclosure) Parent() ComputeModule {
	return ComputeModule{
		Cabinet:       x.Cabinet,
		Chassis:       x.Chassis,
		ComputeModule: x.ComputeModule,
	}
}

// ParentGeneric will determine the parent of this NodeEnclosure, and return it as a GenericXname interface
func (x NodeEnclosure) ParentGeneric() GenericXname {
	return x.Parent()

}

// NodeEnclosurePowerSupply will get a child component with the specified ordinal
func (x NodeEnclosure) NodeEnclosurePowerSupply(nodeEnclosurePowerSupply int) NodeEnclosurePowerSupply {
	return NodeEnclosurePowerSupply{
		Cabinet:                  x.Cabinet,
		Chassis:                  x.Chassis,
		ComputeModule:            x.ComputeModule,
		NodeEnclosure:            x.NodeEnclosure,
		NodeEnclosurePowerSupply: nodeEnclosurePowerSupply,
	}
}

// NodeFpga will get a child component with the specified ordinal
func (x NodeEnclosure) NodeFpga(nodeFpga int) NodeFpga {
	return NodeFpga{
		Cabinet:       x.Cabinet,
		Chassis:       x.Chassis,
		ComputeModule: x.ComputeModule,
		NodeEnclosure: x.NodeEnclosure,
		NodeFpga:      nodeFpga,
	}
}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x NodeEnclosure) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid NodeEnclosure xname: %s", xname)
	}

	return nil
}

// IsController returns whether NodeEnclosure is a controller type, i.e. that
// would host a Redfish entry point
func (x NodeEnclosure) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// NodeEnclosurePowerSupply - xXcCsSbBeEtT
type NodeEnclosurePowerSupply struct {
	Cabinet                  int // xX
	Chassis                  int // cC
	ComputeModule            int // sS
	NodeEnclosure            int // eE
	NodeEnclosurePowerSupply int // tT
}

// Type will return the corresponding HMSType
func (x NodeEnclosurePowerSupply) Type() xnametypes.HMSType {
	return xnametypes.NodeEnclosurePowerSupply
}

// String will stringify NodeEnclosurePowerSupply into the format of xXcCsSbBeEtT
func (x NodeEnclosurePowerSupply) String() string {
	return fmt.Sprintf(
		"x%dc%ds%de%dt%d",
		x.Cabinet,
		x.Chassis,
		x.ComputeModule,
		x.NodeEnclosure,
		x.NodeEnclosurePowerSupply,
	)
}

// Parent will determine the parent of this NodeEnclosurePowerSupply
func (x NodeEnclosurePowerSupply) Parent() NodeEnclosure {
	return NodeEnclosure{
		Cabinet:       x.Cabinet,
		Chassis:       x.Chassis,
		ComputeModule: x.ComputeModule,
		NodeEnclosure: x.NodeEnclosure,
	}
}

// ParentGeneric will determine the parent of this NodeEnclosurePowerSupply, and return it as a GenericXname interface
func (x NodeEnclosurePowerSupply) ParentGeneric() GenericXname {
	return x.Parent()

}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x NodeEnclosurePowerSupply) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid NodeEnclosurePowerSupply xname: %s", xname)
	}

	return nil
}

// IsController returns whether NodeEnclosurePowerSupply is a controller type, i.e. that
// would host a Redfish entry point
func (x NodeEnclosurePowerSupply) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// NodeFpga - xXcCsSbBfF
type NodeFpga struct {
	Cabinet       int // xX
	Chassis       int // cC
	ComputeModule int // sS
	NodeEnclosure int // eE
	NodeFpga      int // fF
}

// Type will return the corresponding HMSType
func (x NodeFpga) Type() xnametypes.HMSType {
	return xnametypes.NodeFpga
}

// String will stringify NodeFpga into the format of xXcCsSbBfF
func (x NodeFpga) String() string {
	return fmt.Sprintf(
		"x%dc%ds%db%df%d",
		x.Cabinet,
		x.Chassis,
		x.ComputeModule,
		x.NodeEnclosure,
		x.NodeFpga,
	)
}

// Parent will determine the parent of this NodeFpga
func (x NodeFpga) Parent() NodeEnclosure {
	return NodeEnclosure{
		Cabinet:       x.Cabinet,
		Chassis:       x.Chassis,
		ComputeModule: x.ComputeModule,
		NodeEnclosure: x.NodeEnclosure,
	}
}

// ParentGeneric will determine the parent of this NodeFpga, and return it as a GenericXname interface
func (x NodeFpga) ParentGeneric() GenericXname {
	return x.Parent()

}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x NodeFpga) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid NodeFpga xname: %s", xname)
	}

	return nil
}

// IsController returns whether NodeFpga is a controller type, i.e. that
// would host a Redfish entry point
func (x NodeFpga) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// NodePowerConnector - xXcCsSv
type NodePowerConnector struct {
	Cabinet            int // xX
	Chassis            int // cC
	ComputeModule      int // sS
	NodePowerConnector int // Sv
}

// Type will return the corresponding HMSType
func (x NodePowerConnector) Type() xnametypes.HMSType {
	return xnametypes.NodePowerConnector
}

// String will stringify NodePowerConnector into the format of xXcCsSv
func (x NodePowerConnector) String() string {
	return fmt.Sprintf(
		"x%dc%ds%dv%d",
		x.Cabinet,
		x.Chassis,
		x.ComputeModule,
		x.NodePowerConnector,
	)
}

// Parent will determine the parent of this NodePowerConnector
func (x NodePowerConnector) Parent() ComputeModule {
	return ComputeModule{
		Cabinet:       x.Cabinet,
		Chassis:       x.Chassis,
		ComputeModule: x.ComputeModule,
	}
}

// ParentGeneric will determine the parent of this NodePowerConnector, and return it as a GenericXname interface
func (x NodePowerConnector) ParentGeneric() GenericXname {
	return x.Parent()

}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x NodePowerConnector) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid NodePowerConnector xname: %s", xname)
	}

	return nil
}

// IsController returns whether NodePowerConnector is a controller type, i.e. that
// would host a Redfish entry point
func (x NodePowerConnector) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// MgmtHLSwitchEnclosure - xXcChH
type MgmtHLSwitchEnclosure struct {
	Cabinet               int // xX
	Chassis               int // cC
	MgmtHLSwitchEnclosure int // hH
}

// Type will return the corresponding HMSType
func (x MgmtHLSwitchEnclosure) Type() xnametypes.HMSType {
	return xnametypes.MgmtHLSwitchEnclosure
}

// String will stringify MgmtHLSwitchEnclosure into the format of xXcChH
func (x MgmtHLSwitchEnclosure) String() string {
	return fmt.Sprintf(
		"x%dc%dh%d",
		x.Cabinet,
		x.Chassis,
		x.MgmtHLSwitchEnclosure,
	)
}

// Parent will determine the parent of this MgmtHLSwitchEnclosure
func (x MgmtHLSwitchEnclosure) Parent() Chassis {
	return Chassis{
		Cabinet: x.Cabinet,
		Chassis: x.Chassis,
	}
}

// ParentGeneric will determine the parent of this MgmtHLSwitchEnclosure, and return it as a GenericXname interface
func (x MgmtHLSwitchEnclosure) ParentGeneric() GenericXname {
	return x.Parent()

}

// MgmtHLSwitch will get a child component with the specified ordinal
func (x MgmtHLSwitchEnclosure) MgmtHLSwitch(mgmtHLSwitch int) MgmtHLSwitch {
	return MgmtHLSwitch{
		Cabinet:               x.Cabinet,
		Chassis:               x.Chassis,
		MgmtHLSwitchEnclosure: x.MgmtHLSwitchEnclosure,
		MgmtHLSwitch:          mgmtHLSwitch,
	}
}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x MgmtHLSwitchEnclosure) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid MgmtHLSwitchEnclosure xname: %s", xname)
	}

	return nil
}

// IsController returns whether MgmtHLSwitchEnclosure is a controller type, i.e. that
// would host a Redfish entry point
func (x MgmtHLSwitchEnclosure) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// MgmtHLSwitch - xXcChHsS
type MgmtHLSwitch struct {
	Cabinet               int // xX
	Chassis               int // cC
	MgmtHLSwitchEnclosure int // hH
	MgmtHLSwitch          int // sS
}

// Type will return the corresponding HMSType
func (x MgmtHLSwitch) Type() xnametypes.HMSType {
	return xnametypes.MgmtHLSwitch
}

// String will stringify MgmtHLSwitch into the format of xXcChHsS
func (x MgmtHLSwitch) String() string {
	return fmt.Sprintf(
		"x%dc%dh%ds%d",
		x.Cabinet,
		x.Chassis,
		x.MgmtHLSwitchEnclosure,
		x.MgmtHLSwitch,
	)
}

// Parent will determine the parent of this MgmtHLSwitch
func (x MgmtHLSwitch) Parent() MgmtHLSwitchEnclosure {
	return MgmtHLSwitchEnclosure{
		Cabinet:               x.Cabinet,
		Chassis:               x.Chassis,
		MgmtHLSwitchEnclosure: x.MgmtHLSwitchEnclosure,
	}
}

// ParentGeneric will determine the parent of this MgmtHLSwitch, and return it as a GenericXname interface
func (x MgmtHLSwitch) ParentGeneric() GenericXname {
	return x.Parent()

}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x MgmtHLSwitch) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid MgmtHLSwitch xname: %s", xname)
	}

	return nil
}

// IsController returns whether MgmtHLSwitch is a controller type, i.e. that
// would host a Redfish entry point
func (x MgmtHLSwitch) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// MgmtSwitch - xXcCwW
type MgmtSwitch struct {
	Cabinet    int // xX
	Chassis    int // cC
	MgmtSwitch int // wW
}

// Type will return the corresponding HMSType
func (x MgmtSwitch) Type() xnametypes.HMSType {
	return xnametypes.MgmtSwitch
}

// String will stringify MgmtSwitch into the format of xXcCwW
func (x MgmtSwitch) String() string {
	return fmt.Sprintf(
		"x%dc%dw%d",
		x.Cabinet,
		x.Chassis,
		x.MgmtSwitch,
	)
}

// Parent will determine the parent of this MgmtSwitch
func (x MgmtSwitch) Parent() Chassis {
	return Chassis{
		Cabinet: x.Cabinet,
		Chassis: x.Chassis,
	}
}

// ParentGeneric will determine the parent of this MgmtSwitch, and return it as a GenericXname interface
func (x MgmtSwitch) ParentGeneric() GenericXname {
	return x.Parent()

}

// MgmtSwitchConnector will get a child component with the specified ordinal
func (x MgmtSwitch) MgmtSwitchConnector(mgmtSwitchConnector int) MgmtSwitchConnector {
	return MgmtSwitchConnector{
		Cabinet:             x.Cabinet,
		Chassis:             x.Chassis,
		MgmtSwitch:          x.MgmtSwitch,
		MgmtSwitchConnector: mgmtSwitchConnector,
	}
}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x MgmtSwitch) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid MgmtSwitch xname: %s", xname)
	}

	return nil
}

// IsController returns whether MgmtSwitch is a controller type, i.e. that
// would host a Redfish entry point
func (x MgmtSwitch) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// MgmtSwitchConnector - xXcCwWjJ
type MgmtSwitchConnector struct {
	Cabinet             int // xX
	Chassis             int // cC
	MgmtSwitch          int // wW
	MgmtSwitchConnector int // jJ
}

// Type will return the corresponding HMSType
func (x MgmtSwitchConnector) Type() xnametypes.HMSType {
	return xnametypes.MgmtSwitchConnector
}

// String will stringify MgmtSwitchConnector into the format of xXcCwWjJ
func (x MgmtSwitchConnector) String() string {
	return fmt.Sprintf(
		"x%dc%dw%dj%d",
		x.Cabinet,
		x.Chassis,
		x.MgmtSwitch,
		x.MgmtSwitchConnector,
	)
}

// Parent will determine the parent of this MgmtSwitchConnector
func (x MgmtSwitchConnector) Parent() MgmtSwitch {
	return MgmtSwitch{
		Cabinet:    x.Cabinet,
		Chassis:    x.Chassis,
		MgmtSwitch: x.MgmtSwitch,
	}
}

// ParentGeneric will determine the parent of this MgmtSwitchConnector, and return it as a GenericXname interface
func (x MgmtSwitchConnector) ParentGeneric() GenericXname {
	return x.Parent()

}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x MgmtSwitchConnector) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid MgmtSwitchConnector xname: %s", xname)
	}

	return nil
}

// IsController returns whether MgmtSwitchConnector is a controller type, i.e. that
// would host a Redfish entry point
func (x MgmtSwitchConnector) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// RouterModule - xXcCrR
type RouterModule struct {
	Cabinet      int // xX
	Chassis      int // cC
	RouterModule int // rR
}

// Type will return the corresponding HMSType
func (x RouterModule) Type() xnametypes.HMSType {
	return xnametypes.RouterModule
}

// String will stringify RouterModule into the format of xXcCrR
func (x RouterModule) String() string {
	return fmt.Sprintf(
		"x%dc%dr%d",
		x.Cabinet,
		x.Chassis,
		x.RouterModule,
	)
}

// Parent will determine the parent of this RouterModule
func (x RouterModule) Parent() Chassis {
	return Chassis{
		Cabinet: x.Cabinet,
		Chassis: x.Chassis,
	}
}

// ParentGeneric will determine the parent of this RouterModule, and return it as a GenericXname interface
func (x RouterModule) ParentGeneric() GenericXname {
	return x.Parent()

}

// HSNAsic will get a child component with the specified ordinal
func (x RouterModule) HSNAsic(hSNAsic int) HSNAsic {
	return HSNAsic{
		Cabinet:      x.Cabinet,
		Chassis:      x.Chassis,
		RouterModule: x.RouterModule,
		HSNAsic:      hSNAsic,
	}
}

// HSNBoard will get a child component with the specified ordinal
func (x RouterModule) HSNBoard(hSNBoard int) HSNBoard {
	return HSNBoard{
		Cabinet:      x.Cabinet,
		Chassis:      x.Chassis,
		RouterModule: x.RouterModule,
		HSNBoard:     hSNBoard,
	}
}

// HSNConnector will get a child component with the specified ordinal
func (x RouterModule) HSNConnector(hSNConnector int) HSNConnector {
	return HSNConnector{
		Cabinet:      x.Cabinet,
		Chassis:      x.Chassis,
		RouterModule: x.RouterModule,
		HSNConnector: hSNConnector,
	}
}

// RouterBMC will get a child component with the specified ordinal
func (x RouterModule) RouterBMC(routerBMC int) RouterBMC {
	return RouterBMC{
		Cabinet:      x.Cabinet,
		Chassis:      x.Chassis,
		RouterModule: x.RouterModule,
		RouterBMC:    routerBMC,
	}
}

// RouterFpga will get a child component with the specified ordinal
func (x RouterModule) RouterFpga(routerFpga int) RouterFpga {
	return RouterFpga{
		Cabinet:      x.Cabinet,
		Chassis:      x.Chassis,
		RouterModule: x.RouterModule,
		RouterFpga:   routerFpga,
	}
}

// RouterPowerConnector will get a child component with the specified ordinal
func (x RouterModule) RouterPowerConnector(routerPowerConnector int) RouterPowerConnector {
	return RouterPowerConnector{
		Cabinet:              x.Cabinet,
		Chassis:              x.Chassis,
		RouterModule:         x.RouterModule,
		RouterPowerConnector: routerPowerConnector,
	}
}

// RouterTOR will get a child component with the specified ordinal
func (x RouterModule) RouterTOR(routerTOR int) RouterTOR {
	return RouterTOR{
		Cabinet:      x.Cabinet,
		Chassis:      x.Chassis,
		RouterModule: x.RouterModule,
		RouterTOR:    routerTOR,
	}
}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x RouterModule) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid RouterModule xname: %s", xname)
	}

	return nil
}

// IsController returns whether RouterModule is a controller type, i.e. that
// would host a Redfish entry point
func (x RouterModule) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// HSNAsic - xXcCrRaA
type HSNAsic struct {
	Cabinet      int // xX
	Chassis      int // cC
	RouterModule int // rR
	HSNAsic      int // aA
}

// Type will return the corresponding HMSType
func (x HSNAsic) Type() xnametypes.HMSType {
	return xnametypes.HSNAsic
}

// String will stringify HSNAsic into the format of xXcCrRaA
func (x HSNAsic) String() string {
	return fmt.Sprintf(
		"x%dc%dr%da%d",
		x.Cabinet,
		x.Chassis,
		x.RouterModule,
		x.HSNAsic,
	)
}

// Parent will determine the parent of this HSNAsic
func (x HSNAsic) Parent() RouterModule {
	return RouterModule{
		Cabinet:      x.Cabinet,
		Chassis:      x.Chassis,
		RouterModule: x.RouterModule,
	}
}

// ParentGeneric will determine the parent of this HSNAsic, and return it as a GenericXname interface
func (x HSNAsic) ParentGeneric() GenericXname {
	return x.Parent()

}

// HSNLink will get a child component with the specified ordinal
func (x HSNAsic) HSNLink(hSNLink int) HSNLink {
	return HSNLink{
		Cabinet:      x.Cabinet,
		Chassis:      x.Chassis,
		RouterModule: x.RouterModule,
		HSNAsic:      x.HSNAsic,
		HSNLink:      hSNLink,
	}
}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x HSNAsic) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid HSNAsic xname: %s", xname)
	}

	return nil
}

// IsController returns whether HSNAsic is a controller type, i.e. that
// would host a Redfish entry point
func (x HSNAsic) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// HSNLink - xXcCrRaAlL
type HSNLink struct {
	Cabinet      int // xX
	Chassis      int // cC
	RouterModule int // rR
	HSNAsic      int // aA
	HSNLink      int // lL
}

// Type will return the corresponding HMSType
func (x HSNLink) Type() xnametypes.HMSType {
	return xnametypes.HSNLink
}

// String will stringify HSNLink into the format of xXcCrRaAlL
func (x HSNLink) String() string {
	return fmt.Sprintf(
		"x%dc%dr%da%dl%d",
		x.Cabinet,
		x.Chassis,
		x.RouterModule,
		x.HSNAsic,
		x.HSNLink,
	)
}

// Parent will determine the parent of this HSNLink
func (x HSNLink) Parent() HSNAsic {
	return HSNAsic{
		Cabinet:      x.Cabinet,
		Chassis:      x.Chassis,
		RouterModule: x.RouterModule,
		HSNAsic:      x.HSNAsic,
	}
}

// ParentGeneric will determine the parent of this HSNLink, and return it as a GenericXname interface
func (x HSNLink) ParentGeneric() GenericXname {
	return x.Parent()

}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x HSNLink) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid HSNLink xname: %s", xname)
	}

	return nil
}

// IsController returns whether HSNLink is a controller type, i.e. that
// would host a Redfish entry point
func (x HSNLink) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// HSNBoard - xXcCrReE
type HSNBoard struct {
	Cabinet      int // xX
	Chassis      int // cC
	RouterModule int // rR
	HSNBoard     int // eE
}

// Type will return the corresponding HMSType
func (x HSNBoard) Type() xnametypes.HMSType {
	return xnametypes.HSNBoard
}

// String will stringify HSNBoard into the format of xXcCrReE
func (x HSNBoard) String() string {
	return fmt.Sprintf(
		"x%dc%dr%de%d",
		x.Cabinet,
		x.Chassis,
		x.RouterModule,
		x.HSNBoard,
	)
}

// Parent will determine the parent of this HSNBoard
func (x HSNBoard) Parent() RouterModule {
	return RouterModule{
		Cabinet:      x.Cabinet,
		Chassis:      x.Chassis,
		RouterModule: x.RouterModule,
	}
}

// ParentGeneric will determine the parent of this HSNBoard, and return it as a GenericXname interface
func (x HSNBoard) ParentGeneric() GenericXname {
	return x.Parent()

}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x HSNBoard) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid HSNBoard xname: %s", xname)
	}

	return nil
}

// IsController returns whether HSNBoard is a controller type, i.e. that
// would host a Redfish entry point
func (x HSNBoard) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// HSNConnector - xXcCrRjJ
type HSNConnector struct {
	Cabinet      int // xX
	Chassis      int // cC
	RouterModule int // rR
	HSNConnector int // jJ
}

// Type will return the corresponding HMSType
func (x HSNConnector) Type() xnametypes.HMSType {
	return xnametypes.HSNConnector
}

// String will stringify HSNConnector into the format of xXcCrRjJ
func (x HSNConnector) String() string {
	return fmt.Sprintf(
		"x%dc%dr%dj%d",
		x.Cabinet,
		x.Chassis,
		x.RouterModule,
		x.HSNConnector,
	)
}

// Parent will determine the parent of this HSNConnector
func (x HSNConnector) Parent() RouterModule {
	return RouterModule{
		Cabinet:      x.Cabinet,
		Chassis:      x.Chassis,
		RouterModule: x.RouterModule,
	}
}

// ParentGeneric will determine the parent of this HSNConnector, and return it as a GenericXname interface
func (x HSNConnector) ParentGeneric() GenericXname {
	return x.Parent()

}

// HSNConnectorPort will get a child component with the specified ordinal
func (x HSNConnector) HSNConnectorPort(hSNConnectorPort int) HSNConnectorPort {
	return HSNConnectorPort{
		Cabinet:          x.Cabinet,
		Chassis:          x.Chassis,
		RouterModule:     x.RouterModule,
		HSNConnector:     x.HSNConnector,
		HSNConnectorPort: hSNConnectorPort,
	}
}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x HSNConnector) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid HSNConnector xname: %s", xname)
	}

	return nil
}

// IsController returns whether HSNConnector is a controller type, i.e. that
// would host a Redfish entry point
func (x HSNConnector) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// HSNConnectorPort - xXcCrRjJpP
type HSNConnectorPort struct {
	Cabinet          int // xX
	Chassis          int // cC
	RouterModule     int // rR
	HSNConnector     int // jJ
	HSNConnectorPort int // pP
}

// Type will return the corresponding HMSType
func (x HSNConnectorPort) Type() xnametypes.HMSType {
	return xnametypes.HSNConnectorPort
}

// String will stringify HSNConnectorPort into the format of xXcCrRjJpP
func (x HSNConnectorPort) String() string {
	return fmt.Sprintf(
		"x%dc%dr%dj%dp%d",
		x.Cabinet,
		x.Chassis,
		x.RouterModule,
		x.HSNConnector,
		x.HSNConnectorPort,
	)
}

// Parent will determine the parent of this HSNConnectorPort
func (x HSNConnectorPort) Parent() HSNConnector {
	return HSNConnector{
		Cabinet:      x.Cabinet,
		Chassis:      x.Chassis,
		RouterModule: x.RouterModule,
		HSNConnector: x.HSNConnector,
	}
}

// ParentGeneric will determine the parent of this HSNConnectorPort, and return it as a GenericXname interface
func (x HSNConnectorPort) ParentGeneric() GenericXname {
	return x.Parent()

}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x HSNConnectorPort) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid HSNConnectorPort xname: %s", xname)
	}

	return nil
}

// IsController returns whether HSNConnectorPort is a controller type, i.e. that
// would host a Redfish entry point
func (x HSNConnectorPort) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// RouterBMC - xXcCrRbB
type RouterBMC struct {
	Cabinet      int // xX
	Chassis      int // cC
	RouterModule int // rR
	RouterBMC    int // bB
}

// Type will return the corresponding HMSType
func (x RouterBMC) Type() xnametypes.HMSType {
	return xnametypes.RouterBMC
}

// String will stringify RouterBMC into the format of xXcCrRbB
func (x RouterBMC) String() string {
	return fmt.Sprintf(
		"x%dc%dr%db%d",
		x.Cabinet,
		x.Chassis,
		x.RouterModule,
		x.RouterBMC,
	)
}

// Parent will determine the parent of this RouterBMC
func (x RouterBMC) Parent() RouterModule {
	return RouterModule{
		Cabinet:      x.Cabinet,
		Chassis:      x.Chassis,
		RouterModule: x.RouterModule,
	}
}

// ParentGeneric will determine the parent of this RouterBMC, and return it as a GenericXname interface
func (x RouterBMC) ParentGeneric() GenericXname {
	return x.Parent()

}

// RouterBMCNic will get a child component with the specified ordinal
func (x RouterBMC) RouterBMCNic(routerBMCNic int) RouterBMCNic {
	return RouterBMCNic{
		Cabinet:      x.Cabinet,
		Chassis:      x.Chassis,
		RouterModule: x.RouterModule,
		RouterBMC:    x.RouterBMC,
		RouterBMCNic: routerBMCNic,
	}
}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x RouterBMC) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid RouterBMC xname: %s", xname)
	}

	return nil
}

// IsController returns whether RouterBMC is a controller type, i.e. that
// would host a Redfish entry point
func (x RouterBMC) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// RouterBMCNic - xXcCrRbBiI
type RouterBMCNic struct {
	Cabinet      int // xX
	Chassis      int // cC
	RouterModule int // rR
	RouterBMC    int // bB
	RouterBMCNic int // iI
}

// Type will return the corresponding HMSType
func (x RouterBMCNic) Type() xnametypes.HMSType {
	return xnametypes.RouterBMCNic
}

// String will stringify RouterBMCNic into the format of xXcCrRbBiI
func (x RouterBMCNic) String() string {
	return fmt.Sprintf(
		"x%dc%dr%db%di%d",
		x.Cabinet,
		x.Chassis,
		x.RouterModule,
		x.RouterBMC,
		x.RouterBMCNic,
	)
}

// Parent will determine the parent of this RouterBMCNic
func (x RouterBMCNic) Parent() RouterBMC {
	return RouterBMC{
		Cabinet:      x.Cabinet,
		Chassis:      x.Chassis,
		RouterModule: x.RouterModule,
		RouterBMC:    x.RouterBMC,
	}
}

// ParentGeneric will determine the parent of this RouterBMCNic, and return it as a GenericXname interface
func (x RouterBMCNic) ParentGeneric() GenericXname {
	return x.Parent()

}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x RouterBMCNic) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid RouterBMCNic xname: %s", xname)
	}

	return nil
}

// IsController returns whether RouterBMCNic is a controller type, i.e. that
// would host a Redfish entry point
func (x RouterBMCNic) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// RouterFpga - xXcCrRfF
type RouterFpga struct {
	Cabinet      int // xX
	Chassis      int // cC
	RouterModule int // rR
	RouterFpga   int // fF
}

// Type will return the corresponding HMSType
func (x RouterFpga) Type() xnametypes.HMSType {
	return xnametypes.RouterFpga
}

// String will stringify RouterFpga into the format of xXcCrRfF
func (x RouterFpga) String() string {
	return fmt.Sprintf(
		"x%dc%dr%df%d",
		x.Cabinet,
		x.Chassis,
		x.RouterModule,
		x.RouterFpga,
	)
}

// Parent will determine the parent of this RouterFpga
func (x RouterFpga) Parent() RouterModule {
	return RouterModule{
		Cabinet:      x.Cabinet,
		Chassis:      x.Chassis,
		RouterModule: x.RouterModule,
	}
}

// ParentGeneric will determine the parent of this RouterFpga, and return it as a GenericXname interface
func (x RouterFpga) ParentGeneric() GenericXname {
	return x.Parent()

}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x RouterFpga) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid RouterFpga xname: %s", xname)
	}

	return nil
}

// IsController returns whether RouterFpga is a controller type, i.e. that
// would host a Redfish entry point
func (x RouterFpga) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// RouterPowerConnector - xXcCrRvV
type RouterPowerConnector struct {
	Cabinet              int // xX
	Chassis              int // cC
	RouterModule         int // rR
	RouterPowerConnector int // vV
}

// Type will return the corresponding HMSType
func (x RouterPowerConnector) Type() xnametypes.HMSType {
	return xnametypes.RouterPowerConnector
}

// String will stringify RouterPowerConnector into the format of xXcCrRvV
func (x RouterPowerConnector) String() string {
	return fmt.Sprintf(
		"x%dc%dr%dv%d",
		x.Cabinet,
		x.Chassis,
		x.RouterModule,
		x.RouterPowerConnector,
	)
}

// Parent will determine the parent of this RouterPowerConnector
func (x RouterPowerConnector) Parent() RouterModule {
	return RouterModule{
		Cabinet:      x.Cabinet,
		Chassis:      x.Chassis,
		RouterModule: x.RouterModule,
	}
}

// ParentGeneric will determine the parent of this RouterPowerConnector, and return it as a GenericXname interface
func (x RouterPowerConnector) ParentGeneric() GenericXname {
	return x.Parent()

}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x RouterPowerConnector) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid RouterPowerConnector xname: %s", xname)
	}

	return nil
}

// IsController returns whether RouterPowerConnector is a controller type, i.e. that
// would host a Redfish entry point
func (x RouterPowerConnector) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// RouterTOR - xXcCrRtT
type RouterTOR struct {
	Cabinet      int // xX
	Chassis      int // cC
	RouterModule int // rR
	RouterTOR    int // tT
}

// Type will return the corresponding HMSType
func (x RouterTOR) Type() xnametypes.HMSType {
	return xnametypes.RouterTOR
}

// String will stringify RouterTOR into the format of xXcCrRtT
func (x RouterTOR) String() string {
	return fmt.Sprintf(
		"x%dc%dr%dt%d",
		x.Cabinet,
		x.Chassis,
		x.RouterModule,
		x.RouterTOR,
	)
}

// Parent will determine the parent of this RouterTOR
func (x RouterTOR) Parent() RouterModule {
	return RouterModule{
		Cabinet:      x.Cabinet,
		Chassis:      x.Chassis,
		RouterModule: x.RouterModule,
	}
}

// ParentGeneric will determine the parent of this RouterTOR, and return it as a GenericXname interface
func (x RouterTOR) ParentGeneric() GenericXname {
	return x.Parent()

}

// RouterTORFpga will get a child component with the specified ordinal
func (x RouterTOR) RouterTORFpga(routerTORFpga int) RouterTORFpga {
	return RouterTORFpga{
		Cabinet:       x.Cabinet,
		Chassis:       x.Chassis,
		RouterModule:  x.RouterModule,
		RouterTOR:     x.RouterTOR,
		RouterTORFpga: routerTORFpga,
	}
}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x RouterTOR) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid RouterTOR xname: %s", xname)
	}

	return nil
}

// IsController returns whether RouterTOR is a controller type, i.e. that
// would host a Redfish entry point
func (x RouterTOR) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}

// RouterTORFpga - xXcCrRtTfF
type RouterTORFpga struct {
	Cabinet       int // xX
	Chassis       int // cC
	RouterModule  int // rR
	RouterTOR     int // tT
	RouterTORFpga int // fF
}

// Type will return the corresponding HMSType
func (x RouterTORFpga) Type() xnametypes.HMSType {
	return xnametypes.RouterTORFpga
}

// String will stringify RouterTORFpga into the format of xXcCrRtTfF
func (x RouterTORFpga) String() string {
	return fmt.Sprintf(
		"x%dc%dr%dt%df%d",
		x.Cabinet,
		x.Chassis,
		x.RouterModule,
		x.RouterTOR,
		x.RouterTORFpga,
	)
}

// Parent will determine the parent of this RouterTORFpga
func (x RouterTORFpga) Parent() RouterTOR {
	return RouterTOR{
		Cabinet:      x.Cabinet,
		Chassis:      x.Chassis,
		RouterModule: x.RouterModule,
		RouterTOR:    x.RouterTOR,
	}
}

// ParentGeneric will determine the parent of this RouterTORFpga, and return it as a GenericXname interface
func (x RouterTORFpga) ParentGeneric() GenericXname {
	return x.Parent()

}

// Validate will validate the string representation of this structure against xnametypes.IsHMSCompIDValid()
func (x RouterTORFpga) Validate() error {
	xname := x.String()
	if !xnametypes.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid RouterTORFpga xname: %s", xname)
	}

	return nil
}

// IsController returns whether RouterTORFpga is a controller type, i.e. that
// would host a Redfish entry point
func (x RouterTORFpga) IsController() bool {
	return xnametypes.IsHMSTypeController(x.Type())
}
