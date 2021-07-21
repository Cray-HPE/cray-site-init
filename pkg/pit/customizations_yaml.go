/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package pit

import (
	"fmt"
	"net"
	"strings"

	"github.com/Cray-HPE/cray-site-init/pkg/csi"
	valid "github.com/asaskevich/govalidator"
	"github.com/spf13/viper"
)

// CustomizationsWLM is the struct for holding all WLM related values for customizations.yaml
type CustomizationsWLM struct {
	WLMStaticIPs struct {
		SlurmCtlD net.IP `yaml:"slurmctld" valid:"ipv4,required" desc:"The slurmCtlD IP address on the nmn, accessible from all UAIs,UANs, and Compute Nodes"`
		SlurmDbd  net.IP `yaml:"slurmdbd" valid:"ipv4,required" desc:"The slurmDbd IP address on the nmn, accessible from all UAIs,UANs, and Compute Nodes"`
		Pbs       net.IP `yaml:"pbs" valid:"ipv4,required" desc:"The PBS IP address on the nmn, accessible from all UAIs,UANs, and Compute Nodes"`
		PbsComm   net.IP `yaml:"pbs_comm" valid:"ipv4,required" desc:"The PBS Comm IP address on the nmn, accessible from all UAIs,UANs, and Compute Nodes"`
	}
	MacVlanSetup struct {
		NMNSubnetCIDR              string `yaml:"nmn_subnet" valid:"cidr,required"`
		NMNSupernetCIDR            string `yaml:"nmn_supernet" valid:"cidr,required"`
		NMNSupernetGateway         net.IP `yaml:"nmn_supernet_gateway" valid:"ipv4,required"`
		NMNVlanInterface           string `yaml:"nmn_vlan" valid:"_,required"`
		NMNMacVlanReservationStart net.IP `yaml:"nmn_reservation_start" valid:"ipv4,required"`
		NMNMacVlanReservationEnd   net.IP `yaml:"nmn_reservation_end" valid:"ipv4,required"`
		Routes                     []struct {
			Destination string `yaml:"dst" valid:"cidr,required"`
			Gateway     string `yaml:"gw" valid:"cidr,required"`
		}
	}
}

// CustomizationsNetworking is the struct for holding all Networking related variables for use in customizations.yaml
type CustomizationsNetworking struct {
	NMN          string `yaml:"nmn" valid:"cidr,required"`
	NMNLB        string `yaml:"nmn_load_balancers" valid:"cidr,required"`
	HMN          string `yaml:"hmn" valid:"cidr,required"`
	HMNLB        string `yaml:"hmn_load_balancers" valid:"cidr,required"`
	HSN          string `yaml:"high_speed" valid:"cidr,required"`
	NetStaticIps struct {
		SiteToSystem       net.IP   `yaml:"site_to_system_lookups" valid:"ipv4,required"`
		SystemToSite       net.IP   `yaml:"system_to_site_lookups" valid:"ipv4,required"`
		NmnTftp            net.IP   `yaml:"nmn_tftp" valid:"ipv4,required"`
		HmnTftp            net.IP   `yaml:"hmn_tftp" valid:"ipv4,required"`
		NmnAPIGateway      net.IP   `yaml:"nmn_api_gw" valid:"ipv4,required"`
		NmnUnbound         net.IP   `yaml:"nmn_unbound" valid:"ipv4,required"`
		NmnAPIGatewayLocal net.IP   `yaml:"nmn_api_gw_local" valid:"ipv4,required"`
		HmnAPIGateway      net.IP   `yaml:"hmn_api_gw" valid:"ipv4,required"`
		NcnMasters         []net.IP `yaml:"nmn_ncn_masters" valid:"required"`
		NcnStorage         []net.IP `yaml:"nmn_ncn_storage" valid:"required"`
	}
	DNS struct {
		ExternalDomain string `yaml:"external" valid:"host,required"`
		ExternalS3     string `yaml:"external_s3" valid:"required,host"`
		ExternalAuth   string `yaml:"external_auth" valid:"required,host"`
		ExternalAPI    string `yaml:"external_api" valid:"required,host"`
		InternalS3     string `yaml:"internal_s3" valid:"required,host"`
		InternalAPI    string `yaml:"internal_api" valid:"required,host"`
	}
}

// CustomizationsYaml is the golang representation of the customizations.yaml stanza we generate
type CustomizationsYaml struct {
	Networking CustomizationsNetworking `yaml:"network"`
	WLM        CustomizationsWLM        `yaml:"wlm"`
}

// ValidationErrors uses govalidator and struct tags to validate and returns errors
func (c *CustomizationsYaml) ValidationErrors() error {
	_, err := valid.ValidateStruct(c)
	return err
}

// IsValid uses govalidator and struct tags to validate and returns a boolean
func (c *CustomizationsYaml) IsValid() bool {
	result, _ := valid.ValidateStruct(c)
	return result
}

// GenCustomizationsYaml generates our configurations.yaml nested struct
func GenCustomizationsYaml(ncns []csi.LogicalNCN, shastaNetworks map[string]*csi.IPV4Network) CustomizationsYaml {
	v := viper.GetViper()
	systemName := v.GetString("system-name")
	siteDomain := v.GetString("site-domain")

	var output CustomizationsYaml
	// nmnMacvlanSubnet, _ := shastaNetworks["NMN"].LookUpSubnet("uai_macvlan")
	var masters []net.IP
	var storage []net.IP
	for _, ncn := range ncns {
		if ncn.Subrole == "Storage" {
			storage = append(storage, ncn.GetIP("NMN"))
		}
		if ncn.Subrole == "Master" {
			masters = append(masters, ncn.GetIP("NMN"))
		}
	}
	nmnLBs, _ := shastaNetworks["NMNLB"].LookUpSubnet("nmn_metallb_address_pool")
	hmnLBs, _ := shastaNetworks["HMNLB"].LookUpSubnet("hmn_metallb_address_pool")
	uaiNet, _ := shastaNetworks["NMN"].LookUpSubnet("uai_macvlan")
	canStaticNet, _ := shastaNetworks["CAN"].LookUpSubnet("can_metallb_static_pool")
	// Normalize the CIDR before using it
	_, uaiNetCIDR, _ := net.ParseCIDR(uaiNet.CIDR.String())
	var customizationsNetworks = CustomizationsNetworking{
		NMN:   shastaNetworks["NMN"].CIDR,
		NMNLB: shastaNetworks["NMNLB"].CIDR,
		HMN:   shastaNetworks["HMN"].CIDR,
		HMNLB: shastaNetworks["HMNLB"].CIDR,
		HSN:   shastaNetworks["HSN"].CIDR,
		NetStaticIps: struct {
			SiteToSystem       net.IP   "yaml:\"site_to_system_lookups\" valid:\"ipv4,required\""
			SystemToSite       net.IP   "yaml:\"system_to_site_lookups\" valid:\"ipv4,required\""
			NmnTftp            net.IP   "yaml:\"nmn_tftp\" valid:\"ipv4,required\""
			HmnTftp            net.IP   "yaml:\"hmn_tftp\" valid:\"ipv4,required\""
			NmnAPIGateway      net.IP   "yaml:\"nmn_api_gw\" valid:\"ipv4,required\""
			NmnUnbound         net.IP   "yaml:\"nmn_unbound\" valid:\"ipv4,required\""
			NmnAPIGatewayLocal net.IP   "yaml:\"nmn_api_gw_local\" valid:\"ipv4,required\""
			HmnAPIGateway      net.IP   "yaml:\"hmn_api_gw\" valid:\"ipv4,required\""
			NcnMasters         []net.IP "yaml:\"nmn_ncn_masters\" valid:\"required\""
			NcnStorage         []net.IP "yaml:\"nmn_ncn_storage\" valid:\"required\""
		}{
			SiteToSystem:       canStaticNet.LookupReservation("external-dns").IPAddress,
			SystemToSite:       net.ParseIP(strings.Split(v.GetString("site-dns"), ",")[0]),
			NmnTftp:            nmnLBs.LookupReservation("cray-tftp").IPAddress,
			HmnTftp:            hmnLBs.LookupReservation("cray-tftp").IPAddress,
			NmnAPIGateway:      nmnLBs.LookupReservation("istio-ingressgateway").IPAddress,
			NmnUnbound:         nmnLBs.LookupReservation("unbound").IPAddress,
			NmnAPIGatewayLocal: nmnLBs.LookupReservation("istio-ingressgateway-local").IPAddress,
			HmnAPIGateway:      hmnLBs.LookupReservation("istio-ingressgateway").IPAddress,
			NcnMasters:         masters,
			NcnStorage:         storage,
		},
		DNS: struct {
			ExternalDomain string "yaml:\"external\" valid:\"host,required\""
			ExternalS3     string "yaml:\"external_s3\" valid:\"required,host\""
			ExternalAuth   string "yaml:\"external_auth\" valid:\"required,host\""
			ExternalAPI    string "yaml:\"external_api\" valid:\"required,host\""
			InternalS3     string "yaml:\"internal_s3\" valid:\"required,host\""
			InternalAPI    string "yaml:\"internal_api\" valid:\"required,host\""
		}{
			ExternalDomain: strings.ToLower(fmt.Sprintf("%s.%s", systemName, siteDomain)),
			ExternalS3:     strings.ToLower(fmt.Sprintf("s3.%s.%s", systemName, siteDomain)),
			ExternalAuth:   strings.ToLower(fmt.Sprintf("auth.%s.%s", systemName, siteDomain)),
			ExternalAPI:    strings.ToLower(fmt.Sprintf("api.%s.%s", systemName, siteDomain)),
			InternalS3:     "rgw-vip.nmn",
			InternalAPI:    "api-gw-service-nmn.local",
		},
	}
	output.Networking = customizationsNetworks
	output.WLM = CustomizationsWLM{
		WLMStaticIPs: struct {
			SlurmCtlD net.IP "yaml:\"slurmctld\" valid:\"ipv4,required\" desc:\"The slurmCtlD IP address on the nmn, accessible from all UAIs,UANs, and Compute Nodes\""
			SlurmDbd  net.IP "yaml:\"slurmdbd\" valid:\"ipv4,required\" desc:\"The slurmDbd IP address on the nmn, accessible from all UAIs,UANs, and Compute Nodes\""
			Pbs       net.IP "yaml:\"pbs\" valid:\"ipv4,required\" desc:\"The PBS IP address on the nmn, accessible from all UAIs,UANs, and Compute Nodes\""
			PbsComm   net.IP "yaml:\"pbs_comm\" valid:\"ipv4,required\" desc:\"The PBS Comm IP address on the nmn, accessible from all UAIs,UANs, and Compute Nodes\""
		}{
			SlurmCtlD: uaiNet.LookupReservation("slurmctld_service").IPAddress,
			SlurmDbd:  uaiNet.LookupReservation("slurmdbd_service").IPAddress,
			Pbs:       uaiNet.LookupReservation("pbs_service").IPAddress,
			PbsComm:   uaiNet.LookupReservation("pbs_comm_service").IPAddress,
		},
		MacVlanSetup: struct {
			NMNSubnetCIDR              string "yaml:\"nmn_subnet\" valid:\"cidr,required\""
			NMNSupernetCIDR            string "yaml:\"nmn_supernet\" valid:\"cidr,required\""
			NMNSupernetGateway         net.IP "yaml:\"nmn_supernet_gateway\" valid:\"ipv4,required\""
			NMNVlanInterface           string "yaml:\"nmn_vlan\" valid:\"_,required\""
			NMNMacVlanReservationStart net.IP "yaml:\"nmn_reservation_start\" valid:\"ipv4,required\""
			NMNMacVlanReservationEnd   net.IP "yaml:\"nmn_reservation_end\" valid:\"ipv4,required\""
			Routes                     []struct {
				Destination string "yaml:\"dst\" valid:\"cidr,required\""
				Gateway     string "yaml:\"gw\" valid:\"cidr,required\""
			}
		}{
			NMNSubnetCIDR:              uaiNetCIDR.String(),
			NMNSupernetGateway:         uaiNet.Gateway,
			NMNSupernetCIDR:            shastaNetworks["NMN"].CIDR,
			NMNVlanInterface:           fmt.Sprintf("vlan%03d", uaiNet.VlanID),
			NMNMacVlanReservationStart: uaiNet.ReservationStart,
			NMNMacVlanReservationEnd:   uaiNet.ReservationEnd,
		},
	}
	for netName, network := range shastaNetworks {
		if strings.HasPrefix(netName, "NMN") {
			if netName != "NMN" {
				output.WLM.MacVlanSetup.Routes = append(output.WLM.MacVlanSetup.Routes, struct {
					Destination string "yaml:\"dst\" valid:\"cidr,required\""
					Gateway     string "yaml:\"gw\" valid:\"cidr,required\""
				}{
					Destination: network.CIDR,
					Gateway:     uaiNet.Gateway.String(),
				})
			}
		}
	}
	return output
}

func init() {
	valid.TagMap["cidr"] = valid.Validator(func(str string) bool {
		_, _, err := net.ParseCIDR(str)
		return err == nil
	})
}
