/*
 * MIT License
 *
 * (C) Copyright 2025 Hewlett Packard Enterprise Development LP
 *
 * Permission is hereby granted, free of charge, to any person obtaining a
 * copy of this software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation
 * the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the
 * Software is furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included
 * in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 * THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 * OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 * ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 * OTHER DEALINGS IN THE SOFTWARE.
 */

package csm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
	"slices"
	"strings"

	"github.com/Cray-HPE/cray-site-init/pkg/csm"
	"github.com/Cray-HPE/cray-site-init/pkg/csm/hms"
	"github.com/Cray-HPE/cray-site-init/pkg/csm/hms/bss"
	"github.com/Cray-HPE/cray-site-init/pkg/csm/hms/sls"
	"github.com/Cray-HPE/cray-site-init/pkg/networking"
	"github.com/Cray-HPE/hms-bss/pkg/bssTypes"
	slsCommon "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
	"github.com/Cray-HPE/hms-xname/xnames"
	"github.com/hashicorp/go-retryablehttp"
)

func getBSSClient(client *bss.UtilsClient) (err error) {
	token, err := csm.GetToken()
	if err != nil {
		err = fmt.Errorf(
			"could not communicate with CSM, failed to fetch the API token because %v",
			err,
		)
		return err
	}
	*client = *bss.NewBSSClient(
		bss.GetBSSBaseURL(),
		retryablehttp.NewClient().StandardClient(),
		token,
	)
	return err
}

func addIPv6BSSBootParameters(slsNetworks *[]slsCommon.Network) (bootParamsCollection []bssTypes.BootParams, err error) {
	for _, network := range *slsNetworks {
		if slices.Contains(
			networksToPatch,
			network.Name,
		) {
			extraProperties, err := sls.UnmarshalNetworkExtraProperties(&network)
			if err != nil {
				return bootParamsCollection, fmt.Errorf(
					"failed to read [%s] network ExtraProperties because %v",
					network.Name,
					err,
				)
			}
			for _, subnetName := range subnetsToPatch {
				subnet, _, err := extraProperties.LookupSubnet(subnetName)
				if err != nil {
					continue
				}
				for _, IPReservation := range subnet.IPReservations {
					var bootParams *bssTypes.BootParams
					if remove {
						bootParams, err = deleteIPv6BSSBootParametersForXName(IPReservation)
						if err != nil {
							return bootParamsCollection, fmt.Errorf(
								"failed to remove IPv6 data from %s's IPAM because %v",
								IPReservation.Name,
								err,
							)
						}
					} else {
						bootParams, err = createIPv6BSSBootParametersForXName(
							&network,
							&extraProperties,
							&subnet,
							IPReservation,
						)
						if err != nil {
							return bootParamsCollection, fmt.Errorf(
								"failed to add IPv6 data to %s's IPAM because %v",
								IPReservation.Name,
								err,
							)
						}
					}
					if bootParams == nil {
						// IPReservation does not have BSS boot parameters.
						continue
					}
					bootParamsCollection = append(
						bootParamsCollection,
						*bootParams,
					)
				}
			}
		}
	}
	return bootParamsCollection, err
}

func createIPv6BSSBootParametersForXName(network *slsCommon.Network, networkProperties *slsCommon.NetworkExtraProperties, subnet *slsCommon.IPSubnet, reservation slsCommon.IPReservation) (bootParams *bssTypes.BootParams, err error) {
	xname, isNodeType, err := hms.NodeTypeXname(reservation.Comment)
	if err != nil {
		return bootParams, fmt.Errorf(
			"failed to parse xname from IPReservation [%v] because %v",
			reservation,
			err,
		)
	}
	if !isNodeType {
		return bootParams, err
	}
	bootParams, err = getBSSBootParams(xname)
	if err != nil {
		return bootParams, fmt.Errorf(
			"failed to get meta-data for %s because %v",
			reservation.Name,
			err,
		)
	}
	if bootParams.CloudInit.MetaData == nil {
		return bootParams, nil
	}
	err = writeToFile(
		fmt.Sprintf(
			"bss-%s-bootparams-backup",
			reservation.Comment,
		),
		bootParams,
	)
	if err != nil {
		return bootParams, fmt.Errorf(
			"failed to write backup BSS data for %s because %v",
			reservation.Name,
			err,
		)
	}
	metaData, err := unmarshalMetadata(bootParams)
	if err != nil {
		return bootParams, fmt.Errorf(
			"failed to read meta-data for %s because %v",
			reservation.Name,
			err,
		)
	}
	networkCIDR6, err := netip.ParsePrefix(networkProperties.CIDR6)
	if err != nil {
		return bootParams, fmt.Errorf(
			"failed to understand %s's CIDR6 value because %v",
			network.Name,
			err,
		)
	}
	for ipamEntry, ipamConfig := range metaData.IPAM {
		if strings.EqualFold(
			network.Name,
			ipamEntry,
		) {
			continue
		}
		if !force && ipamConfig.IP6 != "" {
			fmt.Printf(
				"%s already had an IPv6 address in its BSS bootparameters for the %s network of %s.\n",
				reservation.Name,
				network.Name,
				reservation.IPAddress6,
			)
			forceAlert = true
			continue
		}
		addr6, parseErr := netip.ParseAddr(reservation.IPAddress6.String())
		if parseErr != nil {
			err = fmt.Errorf(
				"failed to parse %s's IPAddress6 string because %v",
				reservation.Name,
				parseErr,
			)
			return bootParams, err
		}
		prefix6 := netip.PrefixFrom(
			addr6,
			networkCIDR6.Bits(),
		)
		ipamConfig.IP6 = prefix6.String()
		metaData.IPAM[ipamEntry] = ipamConfig

		if !force && ipamConfig.Gateway6 != "" {
			fmt.Printf(
				"%s already had an IPv6 address in its BSS bootparameters for the %s network of %s.\n",
				reservation.Name,
				network.Name,
				reservation.IPAddress6,
			)
			forceAlert = true
			continue
		}
		ipamConfig.Gateway6 = subnet.Gateway6.String()
		metaData.IPAM[ipamEntry] = ipamConfig
	}
	err = setBSSMetaData(
		bootParams,
		metaData,
	)
	if err != nil {
		return bootParams, fmt.Errorf(
			"failed to update meta-data in bootparameters because %v",
			err,
		)
	}
	err = writeToFile(
		fmt.Sprintf(
			"bss-%s-bootparams-patched",
			reservation.Comment,
		),
		bootParams,
	)
	if err != nil {
		return bootParams, fmt.Errorf(
			"failed to write patched bootparameters for %s because %v",
			reservation.Name,
			err,
		)
	}

	return bootParams, err
}

func deleteIPv6BSSBootParametersForXName(reservation slsCommon.IPReservation) (bootParams *bssTypes.BootParams, err error) {
	xname, isNodeType, err := hms.NodeTypeXname(reservation.Comment)
	if err != nil {
		return bootParams, fmt.Errorf(
			"failed to parse xname from IPReservation [%v] because %v",
			reservation,
			err,
		)
	}
	if !isNodeType {
		return bootParams, err
	}
	bootParams, err = getBSSBootParams(xname)
	if err != nil {
		return bootParams, fmt.Errorf(
			"failed to get meta-data for %s because %v",
			reservation.Name,
			err,
		)
	}
	err = writeToFile(
		fmt.Sprintf(
			"bss-%s-bootparams-backup",
			reservation.Comment,
		),
		bootParams,
	)
	if err != nil {
		return bootParams, fmt.Errorf(
			"failed to write backup BSS data for %s because %v",
			reservation.Name,
			err,
		)
	}
	metaData, err := unmarshalMetadata(bootParams)
	if err != nil {
		return bootParams, fmt.Errorf(
			"failed to read meta-data for %s (%s) because %v",
			reservation.Name,
			reservation.Comment,
			err,
		)
	}
	for ipamEntry, ipamConfig := range metaData.IPAM {
		ipamConfig.IP6 = ""
		ipamConfig.Gateway6 = ""
		metaData.IPAM[ipamEntry] = ipamConfig
	}
	err = setBSSMetaData(
		bootParams,
		metaData,
	)
	if err != nil {
		return bootParams, fmt.Errorf(
			"failed to update meta-data in bootparameters because %v",
			err,
		)
	}
	err = writeToFile(
		fmt.Sprintf(
			"bss-%s-bootparams-patched",
			reservation.Comment,
		),
		bootParams,
	)
	if err != nil {
		return bootParams, fmt.Errorf(
			"failed to write patched bootparameters for %s (%s) because %v",
			reservation.Name,
			reservation.Comment,
			err,
		)
	}
	return bootParams, err
}

func getBSSBootParams(xname xnames.Xname) (bootParams *bssTypes.BootParams, err error) {
	bootParams, err = bssSession.GetBSSBootparametersForXname(xname.String())
	if err != nil {
		return bootParams, fmt.Errorf(
			"invalid response from BSS for %s because %v",
			xname,
			err,
		)
	}
	return bootParams, err
}

func unmarshalMetadata(bootParams *bssTypes.BootParams) (metaData networking.MetaData, err error) {
	if bootParams.CloudInit.MetaData == nil {
		// not everything will  have BSS bootparameters (e.g. network devices)
		return metaData, nil
	}
	metaDataSerialized, err := json.Marshal(bootParams.CloudInit.MetaData)
	if err != nil {
		return metaData, fmt.Errorf(
			"failed to serialize cloud-init data for parsing because %v",
			err,
		)
	}
	err = json.Unmarshal(
		metaDataSerialized,
		&metaData,
	)
	if err != nil {
		return metaData, fmt.Errorf(
			"failed to read meta-data because %v",
			err,
		)
	}
	return metaData, err
}

func setBSSMetaData(bootParams *bssTypes.BootParams, metaData networking.MetaData) (err error) {
	newMetaDataSerialized, err := json.Marshal(metaData)
	if err != nil {
		return fmt.Errorf(
			"failed to serialize new cloud-init data for parsing because %v",
			err,
		)
	}
	var newMetaData *bssTypes.CloudDataType
	err = json.Unmarshal(
		newMetaDataSerialized,
		&newMetaData,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to read new meta-data because %v",
			err,
		)
	}
	bootParams.CloudInit.MetaData = *newMetaData
	return err
}

func putBSSBootparemeters(client bss.UtilsClient, bootParameterCollection []bssTypes.BootParams) (err error) {
	for _, bootparameters := range bootParameterCollection {
		_, localErr := client.UploadEntryToBSS(
			bootparameters,
			http.MethodPut,
		)
		if localErr != nil {
			err = fmt.Errorf(
				"failed to update %s bootparemeters because %v",
				bootparameters.Hosts[0],
				err,
			)
			break
		}
	}
	return err
}
