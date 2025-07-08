// MIT License
//
// (C) Copyright [2024] Hewlett Packard Enterprise Development LP
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

package rf

import (
	"encoding/json"
	"fmt"
	"strings"
)

const FOXCONN_ETH_INT_DESCRIPTION       = "Foxconn NCSI Interface"
const FOXCONN_PRIMARY_ETH_INT_SUFFIX    = "-primary_eth"
const FOXCONN_PRIMARY_ETH_INT_PCIDID_1  = "0x6315"
const FOXCONN_PRIMARY_ETH_INT_PCIDID_2  = "0x1563"
const FOXCONN_PRIMARY_ETH_INT_FW_NAME   = "X550 FW Ver"

///////////////////////////////////////////////////////////////////////////////
//
// localhost:~ # curl -ks -u root:${BMC_CREDS} https://10.5.1.106/redfish/v1/Systems/system | jq .
// {
//     "@odata.context": "",
//     "@odata.id": "/redfish/v1/Systems/system",
//     "@odata.type": "#ComputerSystem.v1_18_0.ComputerSystem",
//     ... <REDACTED> ...
//     "Oem": {
//         "InsydeNcsi": {
//             "Ncsi": {
//                 "@odata.id": "/redfish/v1/Systems/system/Oem/Insyde/Ncsi"
//             }
//         }
//     }
//     ... <REDACTED> ...
// }
//
///////////////////////////////////////////////////////////////////////////////

type ComputerSystemOemInsyde struct {
	Ncsi ResourceID `json:"Ncsi"`
}

///////////////////////////////////////////////////////////////////////////////
//
// localhost:~ # curl -ks -u root:${BMC_CREDS} https://10.5.1.106/redfish/v1/Systems/system/Oem/Insyde/Ncsi | jq .
// {
//   "@odata.id": "/redfish/v1/Systems/system/Oem/Insyde/Ncsi",
//   "@odata.type": "#InsydeNcsiCollection.InsydeNcsiCollection",
//   "Description": "The NetworkAdapterCollection schema describes a collection of network adapter instances.",
//   "Members": [
//     {
//       "@odata.id": "/redfish/v1/Systems/system/Oem/Insyde/Ncsi/1"
//     },
//     {
//       "@odata.id": "/redfish/v1/Systems/system/Oem/Insyde/Ncsi/2"
//     }
//   ],
//   "Members@odata.count": 2,
//   "Name": "Insyde Ncsi Collection"
// }
//
///////////////////////////////////////////////////////////////////////////////

type InsydeOemNcsiCollection GenericCollection

///////////////////////////////////////////////////////////////////////////////
//
/////// Mellanox Ethernet Interfaces
//
// localhost:~ # curl -ks -u root:${BMC_CREDS} https://10.5.1.106/redfish/v1/Systems/system/Oem/Insyde/Ncsi/1 | jq .
// {
//   "@odata.type": "#InsydeNcsi.v1_0_0.InsydeNcsi",
//   "Description": "The InsydeNcsi schema contains properties related to NCSI device.",
//   "DeviceType": "SMBus",
//   "Id": "1",
//   "Name": "1",
//   "Package": [
//       {
//           "@odata.id": "/redfish/v1/Systems/system/Oem/Insyde/Ncsi/1/Package/0"
//       }
//   ],
//   "VersionID": {
//       "FirmwareName": "mlx0.1",
//       "FirmwareVersion": "1c.28.03.e8",
//       "ManufacturerID": "0x8119",
//       "NcsiVersion": "1.1.0",
//       "PCIDID": "0x1021",
//       "PCISSID": "0x0053",
//       "PCISVID": "0x15b3",
//       "PCIVID": "0x15b3"
//   }
// }
//
/////// Primary Ethernet Interface
//
// localhost:~ # curl -ks -u root:${BMC_CREDS} https://10.5.1.106/redfish/v1/Systems/system/Oem/Insyde/Ncsi/2 | jq .
// {
//   "@odata.id": "/redfish/v1/Systems/system/Oem/Insyde/Ncsi/2",
//   "@odata.type": "#InsydeNcsi.v1_0_0.InsydeNcsi",
//   "Description": "The InsydeNcsi schema contains properties related to NCSI device.",
//   "DeviceType": "NCSIOverRBT",
//   "Id": "2",
//   "Name": "2",
//   "Package": [
//     {
//       "@odata.id": "/redfish/v1/Systems/system/Oem/Insyde/Ncsi/2/Package/1"
//     }
//   ],
//   "VersionID": {
//     "FirmwareName": "X550 FW Ver ",
//     "FirmwareVersion": "00.00.03.60",
//     "ManufacturerID": "0x57010000",
//     "NcsiVersion": "1.0.1",
//     "PCIDID": "0x6315",
//     "PCISSID": "0x0000",
//     "PCIVID": "0x0000"
//   }
// }
//
///////////////////////////////////////////////////////////////////////////////

type InsydeOemNcsiMember struct {
	DeviceType              string                  `json: DeviceType`
	Id                      string                  `json: Id`
	Package                 []ResourceID  			`json:"Package"`
	VersionId               InsydeOemNcsiVersionId  `json:"VersionID"`
}

type InsydeOemNcsiVersionId struct {
	FirmwareName            string                  `json:"FirmwareName"`
	PCIDID                  string                  `json:"PCIDID"`
}

///////////////////////////////////////////////////////////////////////////////
//
// The following are heavily redacted due to size.  Only relevant fields are shown.
//
/////// Mellanox Ethernet Interface
//
// localhost:~ # curl -ks -u root:${BMC_CREDS} https://10.5.1.106/redfish/v1/Systems/system/Oem/Insyde/Ncsi/1/Package/0 | jq .
// {
//   "@odata.id": "/redfish/v1/Systems/system/Oem/Insyde/Ncsi/1/Package/0",
//   "@odata.type": "#InsydeNcsiPackage.v1_0_0.InsydeNcsiPackage",
//   "Description": "The InsydeNcsiPackage schema contains properties related to NcsiPackage.",
//   "Id": "0",
//   "Name": "0",
//   "PackageInfo": [
//     {
//       "ChannelIndex": 0,
//       "MACAddress": "a0:88:c2:7b:17:90",
//     },
//     {
//       "ChannelIndex": 1,
//       "MACAddress": "a0:88:c2:7b:17:91",
//       }
//     }
//   ]
// }
//
/////// Primary Ethernet Interface
//
// localhost:~ # curl -ks -u root:${BMC_CREDS} https://10.5.1.106/redfish/v1/Systems/system/Oem/Insyde/Ncsi/2/Package/1 | jq .
// {
//   "@odata.id": "/redfish/v1/Systems/system/Oem/Insyde/Ncsi/2/Package/1",
//   "@odata.type": "#InsydeNcsiPackage.v1_0_0.InsydeNcsiPackage",
//   "Description": "The InsydeNcsiPackage schema contains properties related to NcsiPackage.",
//   "Id": "1",
//   "Name": "1",
//   "PackageInfo": [
//     {
//       "ChannelIndex": 1,
//       "MACAddress": "04:D9:C8:5D:55:05",
//     },
//     {
//       "ChannelIndex": 2,
//       }
//     }
//   ]
// }
//
///////////////////////////////////////////////////////////////////////////////

type InsydeOemPackage struct {
	Id                      string                  `json: Id`
	PackageInfo             []InsydeOemPackageInfo  `json:"PackageInfo"`
}

type InsydeOemPackageInfo struct {
	ChannelIndex            int                     `json:"ChannelIndex"`
	MACAddress              string                  `json:"MACAddress"`
}

// Parses redfish to find ethernet interfaces for Foxconn Paradise
func discoverFoxconnENetInterfaces(s *EpSystem) {
	//////////////////////////////////////////////////////
	// Parse /redfish/v1/Systems/system/Oem/Insyde/Ncsi

	path := s.SystemRF.OEM.InsydeNcsi.Ncsi.Oid

	url := s.epRF.FQDN + path
	jsonData, err := s.epRF.GETRelative(path)
	if err != nil || jsonData == nil {
		s.LastStatus = HTTPsGetFailed
		errlog.Printf("GET %s failed: %s (jsonData=%s)\n", url, err, jsonData)
		if jsonData == nil {
			errlog.Printf("jsonData is nil\n")
		}
		return
	}
	s.LastStatus = HTTPsGetOk

	var n InsydeOemNcsiCollection
	if err := json.Unmarshal(jsonData, &n); err != nil {
		errlog.Printf("Failed to decode %s: %s\n", url, err)
		s.LastStatus = EPResponseFailedDecode
	}

	if n.MembersOCount < 1 {
		errlog.Printf("No Ncsi members detected")
		return
	}

	//////////////////////////////////////////////////////
	// Parse each /redfish/v1/Systems/system/Oem/Insyde/Ncsi/# entry

	for _, ncsiMember := range n.Members {
		path := ncsiMember.Oid

		url := s.epRF.FQDN + path
		jsonData, err = s.epRF.GETRelative(path)
		if err != nil || jsonData == nil {
			s.LastStatus = HTTPsGetFailed
			errlog.Printf("GET %s failed: %s (jsonData=%s)\n", url, err, jsonData)
			if jsonData == nil {
				errlog.Printf("jsonData is nil\n")
			}
			return
		}
		s.LastStatus = HTTPsGetOk

		var nm InsydeOemNcsiMember
		if err := json.Unmarshal(jsonData, &nm); err != nil {
			errlog.Printf("Failed to decode %s: %s\n", url, err)
			s.LastStatus = EPResponseFailedDecode
		}

		//////////////////////////////////////////////////////
		// Parse each /redfish/v1/Systems/system/Oem/Insyde/Ncsi/# package member element
		// Have only ever seen one package but we should iterate anyway in case that ever changes

		for _, nmPkg := range nm.Package {

			//////////////////////////////////////////////////////
			// Parse /redfish/v1/Systems/system/Oem/Insyde/Ncsi/#/Package/#

			path = nmPkg.Oid

			url = s.epRF.FQDN + path
			jsonData, err = s.epRF.GETRelative(path)
			if err != nil || jsonData == nil {
				s.LastStatus = HTTPsGetFailed
				errlog.Printf("GET %s failed: %s (jsonData=%s)\n", url, err, jsonData)
				if jsonData == nil {
					errlog.Printf("jsonData is nil\n")
				}
				return
			}
			s.LastStatus = HTTPsGetOk

			var p InsydeOemPackage
			if err := json.Unmarshal(jsonData, &p); err != nil {
				errlog.Printf("Failed to decode %s: %s\n", url, err)
				s.LastStatus = EPResponseFailedDecode
			}

			//////////////////////////////////////////////////////
			// Parse each /redfish/v1/Systems/system/Oem/Insyde/Ncsi/#/Package/#.PackageInfo[]

			for j, pi := range p.PackageInfo {
				// Only process if there is a MAC address
				if pi.MACAddress != "" {
					s.ENetInterfaces.Num++

					ei := NewEpEthInterface(s.epRF, ncsiMember.Oid, s.RedfishSubtype, nmPkg, s.ENetInterfaces.Num)

					ei.EtherIfaceRF.Oid = path
					ei.EtherIfaceRF.MACAddress = pi.MACAddress
					ei.EtherIfaceRF.Description = FOXCONN_ETH_INT_DESCRIPTION

					// ID = "ncsi-" + ncsi number + "-p" + package number + "-c" + channel index
					ei.EtherIfaceRF.Id = "ncsi" + nm.Id + "-p" + p.Id + "-c" + fmt.Sprint(j)

					// According to Foxconn, this is the only unique identifier for the primary ethernet
					if nm.VersionId.PCIDID == FOXCONN_PRIMARY_ETH_INT_PCIDID_1 || nm.VersionId.PCIDID == FOXCONN_PRIMARY_ETH_INT_PCIDID_2 {
						// We append a unique string to the end of the ID so that we can identify the
						// primary ethernet interface later.
						ei.EtherIfaceRF.Id += FOXCONN_PRIMARY_ETH_INT_SUFFIX
					} else if strings.TrimSpace(nm.VersionId.FirmwareName) == FOXCONN_PRIMARY_ETH_INT_FW_NAME {
						// Leave a breadcrumb if Foxconn ever changes the PCIDID for the primary ethernet device

						errlog.Printf("Found suspect PCIDID=%s associated with FW \"%s\" for %s\n", nm.VersionId.PCIDID, nm.VersionId.FirmwareName, ei.EtherIfaceRF.Id)

						// TODO: We could append FOXCONN_PRIMARY_ETH_SUFFIX to the ID as well if we (1) think
						// Foxconn will ever change the PCIDID and (2) if we trust FOXCONN_PRIMARY_ETH_INT_FW_NAME
						// will always be unique to the primary ethernet device.  When asked about a unique identifier
						// they said to use the PCIDID.
					}

					ei.BaseOdataID = ei.EtherIfaceRF.Id
					ei.etherIfaceRaw = &jsonData
					ei.LastStatus = VerifyingData

					s.ENetInterfaces.OIDs[ei.EtherIfaceRF.Id] = ei
				}
			}
		}
	}
}

// Determines if this chassis is a Foxconn Paradise chassis.  This is needed
// because Foxconn Paradise chassis can have different manufacturer strings
// and there are times when we haven't yet discovered the manufacturer string
// in the Managers and/or Systems RF endpoints.
func isFoxconnChassis(c *EpChassis) bool {
	// Let's first check the manufacturer string in the Manager for this
	// chassis, if it exists yet.
	var mgr *EpManager = nil
	var ok bool = false

	// Get the first manager linked to
	for _, oid := range c.ManagedBy {
		mgr, ok = c.epRF.Managers.OIDs[oid.Basename()]
		if ok {
			return IsManufacturer(mgr.ManagerRF.Manufacturer, FoxconnMfr) == 1
		}
	}
	// If no link to ManagedBy Manager in the chassis object, just pick
	// the first manager (there is likely only one)
	if !ok {
		for _, m := range c.epRF.Managers.OIDs {
			mgr = m
			return IsManufacturer(mgr.ManagerRF.Manufacturer, FoxconnMfr) == 1
		}
	}

	// The manufacturer string does not yet exist in the Manager for this
	// chassis s0 the fallback is to check if the OdataID of the chassis
	// matches known unique chassis strings for Foxconn Paradise.
	chassisStrings := map[string]struct{} {
		"/redfish/v1/Chassis/Baseboard_0":			{},
		"/redfish/v1/Chassis/BMC_0":				{},
		"/redfish/v1/Chassis/Cpld0":				{},
		"/redfish/v1/Chassis/CPU_0":				{},
		"/redfish/v1/Chassis/CPU_1":				{},
		"/redfish/v1/Chassis/ERoT_CPU_0":			{},
		"/redfish/v1/Chassis/ERoT_CPU_1":			{},
		"/redfish/v1/Chassis/FPGA_0":				{},
		"/redfish/v1/Chassis/Midplane":				{},
		"/redfish/v1/Chassis/NCSI":					{},
		"/redfish/v1/Chassis/PDB":					{},
		"/redfish/v1/Chassis/ProcessorModule_0":	{},
		"/redfish/v1/Chassis/PSU0":					{},
		"/redfish/v1/Chassis/PSU1":					{},
	}

	_, found := chassisStrings[c.OdataID]

	return found
}