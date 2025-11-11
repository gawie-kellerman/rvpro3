package main

import (
	"fmt"
	"log"

	"github.com/gosnmp/gosnmp"
	"github.com/slayercat/GoSNMPServer"
	"github.com/slayercat/GoSNMPServer/mibImps"
)

func main() {
	master := GoSNMPServer.MasterAgent{
		SecurityConfig: GoSNMPServer.SecurityConfig{NoSecurity: true},
		Logger:         GoSNMPServer.NewDefaultLogger(),
		//SecurityConfig: GoSNMPServer.SecurityConfig{
		//	AuthoritativeEngineBoots: 1,
		//	Users: []gosnmp.UsmSecurityParameters{
		//		{
		//			UserName:                 "v3Username",
		//			AuthenticationProtocol:   gosnmp.MD5,
		//			PrivacyProtocol:          gosnmp.DES,
		//			AuthenticationPassphrase: "v3AuthenticationPassphrase",
		//			PrivacyPassphrase:        "v3PrivacyPassphrase",
		//		},
		//	},
		//},
		SubAgents: []*GoSNMPServer.SubAgent{
			{
				CommunityIDs: []string{"community"},
				OIDs:         mibImps.All(),
			},
		},
	}

	server := GoSNMPServer.NewSNMPServer(master)
	err := server.ListenUDP("udp", "127.0.0.1:1161")
	if err != nil {
		log.Fatal("Error in listen: %+v", err)
	}

	value := &GoSNMPServer.PDUValueControlItem{
		OID:               "1.3.6.1.4.1.2021.11.66.0",
		Type:              gosnmp.Counter32,
		NonWalkable:       false,
		OnCheckPermission: nil,
		OnGet:             onGetPDUValue,
		OnSet:             nil,
		OnTrap:            nil,
		Document:          "ifIndex",
	}
	master.SubAgents[0].OIDs = append(master.SubAgents[0].OIDs, value)
	server.ServeForever()
}

func onTrap(inform bool, trapdata gosnmp.SnmpPDU) (dataret interface{}, err error) {
	return 999, nil
}

func onSetValue(value interface{}) error {
	fmt.Println("OnSetValue:", value)
	return nil
}

func onGetPDUValue() (value interface{}, err error) {
	//return GoSNMPServer.Asn1IntegerWrap(999), nil
	return GoSNMPServer.Asn1Counter32Wrap(999), nil
	//return 999, nil
}

func onCheckPermissionCallback(version gosnmp.SnmpVersion, pduType gosnmp.PDUType, name string) GoSNMPServer.PermissionAllowance {
	return GoSNMPServer.PermissionAllowanceAllowed
}
