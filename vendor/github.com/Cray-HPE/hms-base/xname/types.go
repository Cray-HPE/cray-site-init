package xname

import (
	"fmt"

	base "github.com/Cray-HPE/hms-base"
)

// s0
type System struct{}

func (s System) String() string {
	return "s0"
}

func (s System) Validate() error {
	return nil
}

func (s System) ValidateEnhanced() error {
	return nil
}

func (s System) CDU(coolingGroup int) CDU {
	return CDU{
		CoolingGroup: coolingGroup,
	}
}

func (s System) Cabinet(cabinet int) Cabinet {
	return Cabinet{
		Cabinet: cabinet,
	}
}

// dD
type CDU struct {
	CoolingGroup int // D: 0-999
}

func (c CDU) String() string {
	formatStr, _, _ := base.GetHMSTypeFormatString(base.CDU)
	return fmt.Sprintf(formatStr, c.CoolingGroup)
}

func (c CDU) Parent() System {
	return System{}
}

func (c CDU) Validate() error {
	xname := c.String()
	if !base.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid CDU xname: %s", xname)
	}

	return nil
}

func (c CDU) ValidateEnhanced() error {
	// Perform normal validation
	if c.Validate() != nil {
		// Xname is not valid
	}

	if c.Parent().ValidateEnhanced() != nil {
		// Verify all parents are valid 
	}

	if !(0 <= c.CoolingGroup || c.CoolingGroup <= 999) {
		// Cooling group range
	}

	return nil
}

func (c CDU) CDUMgmtSwitch(slot int) CDUMgmtSwitch {
	return CDUMgmtSwitch{
		CoolingGroup: c.CoolingGroup,
		Slot:         slot,
	}
}

// dDwW
type CDUMgmtSwitch struct {
	CoolingGroup int // D: 0-999
	Slot         int // W: 0-31
}

func (cms CDUMgmtSwitch) String() string {
	formatStr, _, _ := base.GetHMSTypeFormatString(base.CDUMgmtSwitch)
	return fmt.Sprintf(formatStr, cms.CoolingGroup, cms.Slot)
}

func (cms CDUMgmtSwitch) Parent() CDU {
	return CDU{
		CoolingGroup: cms.CoolingGroup,
	}
}

func (cms CDUMgmtSwitch) Validate() error {
	xname := cms.String()
	if !base.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid CDUMgmtSwitch xname: %s", xname)
	}

	return nil
}

func (cms CDUMgmtSwitch) ValidateEnhanced() error {
	// Perform normal validation
	if cms.Validate() != nil {
		// Xname is not valid
	}

	if cms.Parent().ValidateEnhanced() != nil {
		// Verify all parents are valid 
	}

	if !(0 <= cms.Slot || cms.Slot <= 31) {
		// CDU Switch slot
	}

	return nil
}

// xX
type Cabinet struct {
	Cabinet int // X: 0-999
}

func (c Cabinet) String() string {
	formatStr, _, _ := base.GetHMSTypeFormatString(base.Cabinet)
	return fmt.Sprintf(formatStr, c.Cabinet)
}

func (c Cabinet) Parent() System {
	return System{}
}

func (c Cabinet) Validate() error {
	xname := c.String()
	if !base.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid Cabinet xname: %s", xname)
	}

	return nil
}

func (c Cabinet) ValidateEnhanced() error {
	// Perform normal validation
	if c.Validate() != nil {
		// Xname is not valid
	}

	if c.Parent().ValidateEnhanced() != nil {
		// Verify all parents are valid 
	}

	if !(0 <= c.Cabinet && c.Cabinet <= 999) {
		// Cabinet number out of range
	}

	return nil
}

func (c Cabinet) Chassis(chassis int) Chassis {
	return Chassis{
		Cabinet: c.Cabinet,
		Chassis: chassis,
	}
}

func (c Cabinet) CabinetPDUController(pduController int) CabinetPDUController {
	return CabinetPDUController{
		Cabinet:       c.Cabinet,
		PDUController: pduController,
	}
}

// xXmM
type CabinetPDUController struct {
	Cabinet       int // X: 0-999
	PDUController int // M: 0-3
}

func (p CabinetPDUController) String() string {
	formatStr, _, _ := base.GetHMSTypeFormatString(base.CabinetPDUController)
	return fmt.Sprintf(formatStr, p.Cabinet, p.Cabinet)
}

func (p CabinetPDUController) Parent() Cabinet {
	return Cabinet{
		Cabinet: p.Cabinet,
	}
}

func (p CabinetPDUController) Validate() error {
	xname := p.String()
	if !base.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid CabinetPDUController xname: %s", xname)
	}

	return nil
}

func (p CabinetPDUController) ValidateEnhanced() error {
	// Perform normal validation
	if p.Validate() != nil {
		// Xname is not valid
	}

	if p.Parent().ValidateEnhanced() != nil {
		// Verify all parents are valid 
	}

	if !(0 <= p.PDUController && p.PDUController <= 3) {
		// Cabinet number out of range
	}

	return nil
}

// xXcC
// Mountain Have 8 c0-c8
// Hill have 2 c1 and c3
// River always have c0
type Chassis struct {
	Cabinet int // X: 0-999
	Chassis int // C: 0-7
}

func (c Chassis) String() string {
	formatStr, _, _ := base.GetHMSTypeFormatString(base.Chassis)
	return fmt.Sprintf(formatStr, c.Cabinet, c.Chassis)
}

func (c Chassis) Parent() Cabinet {
	return Cabinet{
		Cabinet: c.Cabinet,
	}
}

func (c Chassis) Validate() error {
	xname := c.String()
	if !base.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid Chassis xname: %s", xname)
	}

	return nil
}

func (c Chassis) ValidateEnhanced(class base.HMSClass) error {
	// Perform normal validation
	if c.Validate() != nil {
		// Xname is not valid
	}

	if c.Parent().ValidateEnhanced() != nil {
		// Verify all parents are valid 
	}

	// Chassis Validation
	switch class {
	case base.ClassRiver:
		if c.Chassis != 0 {
			// River chassis must be equal to 0
		}
	case base.ClassHill:
		if !(c.Chassis == 1 || c.Chassis == 3) {
			// Hill has Chassis 1 or 3
		}
	case base.ClassMountain:
		if !(0 <= c.Chassis && c.Chassis <= 7) {
			// Mountain must chassis between 0 and 7
		}
	default:
	}

	return nil
}

func (c Chassis) MgmtHLSwitchEnclosure(slot int) MgmtHLSwitchEnclosure {
	return MgmtHLSwitchEnclosure{
		Cabinet: c.Cabinet,
		Chassis: c.Chassis,
		Slot:    slot,
	}
}

func (c Chassis) MgmtSwitch(slot int) MgmtSwitch {
	return MgmtSwitch{
		Cabinet: c.Cabinet,
		Chassis: c.Chassis,
		Slot:    slot,
	}
}

// This is a convience function, as we normally do not work with MgmtHLSwitchEnclosures directly
func (c Chassis) MgmtHLSwitch(slot, space int) MgmtHLSwitch {
	return c.MgmtHLSwitchEnclosure(slot).MgmtHLSwitch(space)
}

func (c Chassis) RouterModule(slot int) RouterModule {
	return RouterModule{
		Cabinet: c.Cabinet,
		Chassis: c.Chassis,
		Slot:    slot,
	}
}

// This is a convince function, as we normally do not work with RouterModules directly.
func (c Chassis) RouterBMC(slot, bmc int) RouterBMC {
	return c.RouterModule(slot).RouterBMC(bmc)
}

func (c Chassis) ComputeModule(slot int) ComputeModule {
	return ComputeModule{
		Cabinet: c.Cabinet,
		Chassis: c.Chassis,
		Slot:    slot,
	}
}

func (c Chassis) NodeBMC(slot, bmc int) NodeBMC {
	return c.ComputeModule(slot).NodeBMC(bmc)
}

// xXcCwW
type MgmtSwitch struct {
	Cabinet int // X: 0-999
	Chassis int // C: 0-7
	Slot    int // W: 1-48
}

func (ms MgmtSwitch) String() string {
	formatStr, _, _ := base.GetHMSTypeFormatString(base.MgmtSwitch)
	return fmt.Sprintf(formatStr, ms.Cabinet, ms.Chassis, ms.Slot)
}

func (ms MgmtSwitch) Parent() Chassis {
	return Chassis{
		Cabinet: ms.Cabinet,
		Chassis: ms.Chassis,
	}
}

func (ms MgmtSwitch) Validate() error {
	xname := ms.String()
	if !base.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid MgmtSwitch xname: %s", xname)
	}

	return nil
}

func (ms MgmtSwitch) ValidateEnhanced(class base.HMSClass) error {
	// Perform normal validation
	if ms.Validate() != nil {
		// Xname is not valid
	}

	if ms.Parent().ValidateEnhanced(class) != nil {
		// Verify all parents are valid 
	}

	// Chassis Validation
	switch class {
	case base.ClassRiver:
		// Expected to be river only

		if !(1 <= ms.Slot && ms.Slot <= 48) {
			// Verify that the U is within a standard rack
		}
	case base.ClassHill:
		fallthrough
	case base.ClassMountain:
		// MgmtSwitches are only for river
	default:
	}

	return nil
}

func (ms MgmtSwitch) MgmtSwitchConnector(switchPort int) MgmtSwitchConnector {
	return MgmtSwitchConnector{
		Cabinet:    ms.Cabinet,
		Chassis:    ms.Chassis,
		Slot:       ms.Slot,
		SwitchPort: switchPort,
	}
}

// xXcCwWjJ
type MgmtSwitchConnector struct {
	Cabinet    int // X: 0-999
	Chassis    int // C: 0-7
	Slot       int // W: 1-48
	SwitchPort int // J: 1-32 // TODO the HSOS page, should allow upto at least 48
}

func (msc MgmtSwitchConnector) String() string {
	formatStr, _, _ := base.GetHMSTypeFormatString(base.MgmtSwitchConnector)
	return fmt.Sprintf(formatStr, msc.Cabinet, msc.Chassis, msc.Slot, msc.SwitchPort)
}

func (msc MgmtSwitchConnector) Parent() MgmtSwitch {
	return MgmtSwitch{
		Cabinet: msc.Cabinet,
		Chassis: msc.Chassis,
		Slot:    msc.Slot,
	}
}

func (msc MgmtSwitchConnector) Validate() error {
	xname := msc.String()
	if !base.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid MgmtSwitchConnector xname: %s", xname)
	}

	return nil
}

func (msc MgmtSwitchConnector) ValidateEnhanced(class base.HMSClass) error {
	// Perform normal validation
	if msc.Validate() != nil {
		// Xname is not valid
	}

	if msc.Parent().ValidateEnhanced(class) != nil {
		// Verify all parents are valid 
	}

	// Chassis Validation
	switch class {
	case base.ClassRiver:
		// Expected to be river only

		if !(1 <= msc.SwitchPort) {
			// Verify that the switch port is valid
		}
	case base.ClassHill:
		fallthrough
	case base.ClassMountain:
		// MgmtSwitches are only for river
	default:
	}

	return nil
}

// xXcChH
type MgmtHLSwitchEnclosure struct {
	Cabinet int // X: 0-999
	Chassis int // C: 0-7
	Slot    int // H: 1-48
}

func (enclosure MgmtHLSwitchEnclosure) String() string {
	formatStr, _, _ := base.GetHMSTypeFormatString(base.MgmtHLSwitchEnclosure)
	return fmt.Sprintf(formatStr, enclosure.Cabinet, enclosure.Chassis, enclosure.Slot)
}

func (enclosure MgmtHLSwitchEnclosure) Parent() Chassis {
	return Chassis{
		Cabinet: enclosure.Cabinet,
		Chassis: enclosure.Chassis,
	}
}

func (enclosure MgmtHLSwitchEnclosure) Validate() error {
	xname := enclosure.String()
	if !base.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid MgmtHLSwitchEnclosure xname: %s", xname)
	}

	return nil
}

func (enclosure MgmtHLSwitchEnclosure) ValidateEnhanced(class base.HMSClass) error {
	// Perform normal validation
	if enclosure.Validate() != nil {
		// Xname is not valid
	}

	if enclosure.Parent().ValidateEnhanced(class) != nil {
		// Verify all parents are valid 
	}

	// Chassis Validation
	switch class {
	case base.ClassRiver:
		// Expected to be river only

		if !(1 <= enclosure.Slot && enclosure.Slot <= 48) {
			// Verify that the U is within a standard rack
		}
	case base.ClassHill:
		fallthrough
	case base.ClassMountain:
		// MgmtSwitches are only for river
	default:
	}

	return nil
}

func (enclosure MgmtHLSwitchEnclosure) MgmtHLSwitch(space int) MgmtHLSwitch {
	return MgmtHLSwitch{
		Cabinet: enclosure.Cabinet,
		Chassis: enclosure.Chassis,
		Slot:    enclosure.Slot,
		Space:   space,
	}
}

//xXcChHsS
type MgmtHLSwitch struct {
	Cabinet int // X: 0-999
	Chassis int // C: 0-7
	Slot    int // H: 1-48
	Space   int // S: 1-4
}

func (mhls MgmtHLSwitch) String() string {
	formatStr, _, _ := base.GetHMSTypeFormatString(base.MgmtHLSwitch)
	return fmt.Sprintf(formatStr, mhls.Cabinet, mhls.Chassis, mhls.Slot, mhls.Space)
}

func (mhls MgmtHLSwitch) Parent() MgmtHLSwitchEnclosure {
	return MgmtHLSwitchEnclosure{
		Cabinet: mhls.Cabinet,
		Chassis: mhls.Chassis,
		Slot:    mhls.Slot,
	}
}

func (mhls MgmtHLSwitch) Validate() error {
	xname := mhls.String()
	if !base.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid MgmtHLSwitch xname: %s", xname)
	}

	return nil
}

func (mhls MgmtHLSwitch) ValidateEnhanced(class base.HMSClass) error {
	// Perform normal validation
	if mhls.Validate() != nil {
		// Xname is not valid
	}

	if mhls.Parent().ValidateEnhanced(class) != nil {
		// Verify all parents are valid 
	}

	// Chassis Validation
	switch class {
	case base.ClassRiver:
		// Expected to be river only

		if !(1 <= mhls.Space && mhls.Space <= 4) {
			// Verify a valid space value
		}
	case base.ClassHill:
		fallthrough
	case base.ClassMountain:
		// MgmtSwitches are only for river
	default:
	}

	return nil
}

// xXcCrR
// Mountain/Hill: R: 0-8
// River: 1-48
type RouterModule struct {
	Cabinet int // X: 0-999
	Chassis int // C: 0-7
	Slot    int // R: 0-64
}

func (rm RouterModule) String() string {
	formatStr, _, _ := base.GetHMSTypeFormatString(base.RouterModule)
	return fmt.Sprintf(formatStr, rm.Cabinet, rm.Chassis, rm.Slot)
}

func (rm RouterModule) Parent() Chassis {
	return Chassis{
		Cabinet: rm.Cabinet,
		Chassis: rm.Chassis,
	}
}

func (rm RouterModule) Validate() error {
	xname := rm.String()
	if !base.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid RouterModule xname: %s", xname)
	}

	return nil
}

func (rm RouterModule) ValidateEnhanced(class base.HMSClass) error {
	// Perform normal validation
	if rm.Validate() != nil {
		// Xname is not valid
	}

	if rm.Parent().ValidateEnhanced(class) != nil {
		// Verify all parents are valid 
	}

	// Router Module Validation
	switch class {
	case base.ClassRiver:
		if !(1 <= rm.Slot && rm.Slot <= 48) {
			// Standard Rack size			
		}
	case base.ClassHill:
		fallthrough
	case base.ClassMountain:
		if !(0 <= rm.Slot && rm.Slot <= 7) {
			// Mountain Chassis			
		}
	}

	return nil
}

func (rm RouterModule) RouterBMC(bmc int) RouterBMC {
	return RouterBMC{
		Cabinet: rm.Cabinet,
		Chassis: rm.Chassis,
		Slot:    rm.Slot,
		BMC:     bmc,
	}
}

// xXcCrRbB
// B is always 0
type RouterBMC struct {
	Cabinet int // X: 0-999
	Chassis int // C: 0-7
	Slot    int // R: 0-64
	BMC     int // B: 0
}

func (bmc RouterBMC) String() string {
	formatStr, _, _ := base.GetHMSTypeFormatString(base.RouterBMC)
	return fmt.Sprintf(formatStr, bmc.Cabinet, bmc.Chassis, bmc.Slot, bmc.BMC)
}

func (bmc RouterBMC) Parent() RouterModule {
	return RouterModule{
		Cabinet: bmc.Cabinet,
		Chassis: bmc.Chassis,
		Slot:    bmc.Slot,
	}
}

func (bmc RouterBMC) Validate() error {
	xname := bmc.String()
	if !base.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid RouterBMC xname: %s", xname)
	}

	return nil
}

func (bmc RouterBMC) ValidateEnhanced(class base.HMSClass) error {
	// Perform normal validation
	if bmc.Validate() != nil {
		// Xname is not valid
	}

	if bmc.Parent().ValidateEnhanced(class) != nil {
		// Verify all parents are valid 
	}

	// Router BMC Validation
	if bmc.BMC != 0 {
		// BMC should always be 0
	}

	return nil
}

// xXcCsS
// Mountain/Hill: 0-7
// River: 1-48
type ComputeModule struct {
	Cabinet int // X: 0-999
	Chassis int // C: 0-7
	Slot    int // S: 1-63
}

func (cm ComputeModule) String() string {
	formatStr, _, _ := base.GetHMSTypeFormatString(base.ComputeModule)
	return fmt.Sprintf(formatStr, cm.Cabinet, cm.Chassis, cm.Slot)
}

func (cm ComputeModule) Parent() Chassis {
	return Chassis{
		Cabinet: cm.Cabinet,
		Chassis: cm.Chassis,
	}
}

func (cm ComputeModule) Validate() error {
	xname := cm.String()
	if !base.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid ComputeModule xname: %s", xname)
	}

	return nil
}

func (cm ComputeModule) ValidateEnhanced(class base.HMSClass) error {
	// Perform normal validation
	if cm.Validate() != nil {
		// Xname is not valid
	}

	if cm.Parent().ValidateEnhanced(class) != nil {
		// Verify all parents are valid 
	}

	// Compute Module Validation
	switch class {
	case base.ClassRiver:
		if !(1 <= cm.Slot && cm.Slot <= 48) {
			// Standard Rack size			
		}
	case base.ClassHill:
		fallthrough
	case base.ClassMountain:
		if !(0 <= cm.Slot && cm.Slot <= 7) {
			// Mountain Chassis			
		}
	}

	return nil
}

func (cm ComputeModule) NodeBMC(bmc int) NodeBMC {
	return NodeBMC{
		Cabinet: cm.Cabinet,
		Chassis: cm.Chassis,
		Slot:    cm.Slot,
		BMC:     bmc,
	}
}

// xXcCsSbB
// Node Card/Node BMC
// Mountain/Hill can be 0 or 1
// River
// - Single node chassis: always 0
// - Dual Node chassis: 1 or 2
// - Dense/Quad Node Chassis 1-4
type NodeBMC struct {
	Cabinet int // X: 0-999
	Chassis int // C: 0-7
	Slot    int // S: 1-63
	BMC     int // B: 0-1 - TODO the HSOS document is wrong here. as we do actually use greater than 1
}

func (bmc NodeBMC) Parent() ComputeModule {
	return ComputeModule{
		Cabinet: bmc.Cabinet,
		Chassis: bmc.Chassis,
		Slot:    bmc.Slot,
	}
}

func (bmc NodeBMC) Validate() error {
	xname := bmc.String()
	if !base.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid NodeBMC xname: %s", xname)
	}

	return nil
}

func (bmc NodeBMC) ValidateEnhanced(class base.HMSClass, nodeChassisType NodeBladeType) error {
	// Perform normal validation
	if bmc.Validate() != nil {
		// Xname is not valid
	}

	if bmc.Parent().ValidateEnhanced(class) != nil {
		// Verify all parents are valid 
	}

	// NodeBMC Validation
	switch class {
	case base.ClassRiver:
		switch nodeChassisType {
		case SingleNodeBlade:
			if bmc.BMC != 0 {
				// Single node chassis must have the BMC as 0
			}
		case DualNodeBlade:
			if !(bmc.BMC == 1 || bmc.BMC == 2) {
				// Dual node chassis must have BMC as 1 or 2
			}
		case DenseQuadNodeBlade:
			if !(1 <= bmc.BMC || bmc.BMC <= 4) {
				// Dense Quad node chassis must have BMC between 1 and 4
			}
		}
	case base.ClassHill:
		fallthrough
	case base.ClassMountain:
		if !(bmc.BMC == 0 || bmc.BMC == 1) {
			// Mountain blades have 2 BMCs
		}
	default:
	}
	return nil
}

func (bmc NodeBMC) Node(node int) Node {
	return Node{
		Cabinet: bmc.Cabinet,
		Chassis: bmc.Chassis,
		Slot:    bmc.Slot,
		BMC:     bmc.BMC,
		Node:    node,
	}
}

func (bmc NodeBMC) String() string {
	formatStr, _, _ := base.GetHMSTypeFormatString(base.NodeBMC)
	return fmt.Sprintf(formatStr, bmc.Cabinet, bmc.Chassis, bmc.Slot, bmc.BMC)
}

// xCcCsSbBnN
// River - Always 0
// Mountain/Hill 0 or 1
type Node struct {
	Cabinet int // X: 0-999
	Chassis int // C: 0-7
	Slot    int // S: 1-63
	BMC     int // B: 0-1 - TODO the HSOS document is wrong here. as we do actually use greater than 1
	Node    int // N: 0-7
}

func (n Node) String() string {
	formatStr, _, _ := base.GetHMSTypeFormatString(base.Node)
	return fmt.Sprintf(formatStr, n.Cabinet, n.Chassis, n.Slot, n.BMC, n.Node)
}

func (n Node) Validate() error {
	xname := n.String()
	if !base.IsHMSCompIDValid(xname) {
		return fmt.Errorf("invalid node xname: %s", xname)
	}

	return nil
}

type NodeBladeType int // TODO Idk if this should be blade or chassis. This this could apply to both river and mountain

const (
	SingleNodeBlade NodeBladeType = iota
	DualNodeBlade
	DenseQuadNodeBlade // TODO Should this have "dense"
)

func (n Node) ValidateEnhanced(class base.HMSClass, nodeChassisType NodeBladeType) error {
	// Perform normal validation
	if n.Validate() != nil {
		// Xname is not valid
	}

	if n.Parent().ValidateEnhanced(class, nodeChassisType) != nil {
		// Verify all parents are valid 
	}
	
	// Node Validation
	switch class {
	case base.ClassRiver:
		if n.Node != 0 {
			// River node value must be 0
		}
	case base.ClassHill:
		fallthrough
	case base.ClassMountain:
		switch nodeChassisType {
		case SingleNodeBlade:
			// We don't have this?
		case DualNodeBlade:
			if n.Node != 0 {
				// On a mountain dual node blade, each BMC controls 1 node.
			}
		case DenseQuadNodeBlade:
			if !(n.Node == 0 || n.Node == 1) {
				// Dual node blade must have BMC as 1 or 2
			}
		}
	}

	return nil
}

func (n Node) Parent() NodeBMC {
	return NodeBMC{
		Cabinet: n.Cabinet,
		Chassis: n.Chassis,
		Slot:    n.Slot,
		BMC:     n.BMC,
	}
}
