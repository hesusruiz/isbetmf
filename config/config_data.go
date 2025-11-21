package config

import "github.com/hesusruiz/isbetmf/types"

// As this PDP is designed for DOME and ISBE environments, many config data items are configured in code.
// This avoids many configuration errors and simplifies deployment, at the expense of some flexibility.
// However, this flexibility is not really needed in practice, as the DOME environments are well defined and stable.
// Minimizing errors is here much more important than the ease to configure these parameters.

var domedevConfig = &Config{
	Environment:  DOME_DEV,
	ProxyEnabled: true,

	ServerOperatorOrganizationIdentifier: "VATSB-12345678J",
	ServerOperatorDid:                    "did:elsi:VATSB-12345678J",
	ServerOperatorName:                   "DOME Foundation SBX",
	ServerOperatorCountry:                "ES",

	LEARPower: types.OnePower{
		Type:     "domain",
		Domain:   "DOME",
		Function: "Onboarding",
		Action:   []string{"execute"},
	},
	ProductCreatePower: types.OnePower{
		Type:     "domain",
		Domain:   "DOME",
		Function: "ProductOffering",
		Action:   []string{"Create"},
	},
	ProductUpdatePower: types.OnePower{
		Type:     "domain",
		Domain:   "DOME",
		Function: "ProductOffering",
		Action:   []string{"Update"},
	},
	ProductDeletePower: types.OnePower{
		Type:     "domain",
		Domain:   "DOME",
		Function: "ProductOffering",
		Action:   []string{"Delete"},
	},

	PolicyFileName:  "auth_policies.star",
	RemoteTMFServer: "https://tmf.dome-marketplace-sbx.org",
	VerifierServer:  "https://verifier.dome-marketplace-sbx.org",
	Dbname:          "data/tmf.dome.sbx.db",
	ClonePeriod:     DefaultClonePeriod,
}

var isbepreConfig = &Config{
	Environment:  ISBE_PRE,
	ProxyEnabled: false,

	ServerOperatorOrganizationIdentifier: "VATES-G87936159",
	ServerOperatorDid:                    "did:elsi:VATES-G87936159",
	ServerOperatorName:                   "Alastria",
	ServerOperatorCountry:                "ES",

	LEARPower: types.OnePower{
		Type:     "domain",
		Domain:   "DOME",
		Function: "Onboarding",
		Action:   []string{"execute"},
	},
	ProductCreatePower: types.OnePower{
		Type:     "domain",
		Domain:   "DOME",
		Function: "ProductOffering",
		Action:   []string{"Create"},
	},
	ProductUpdatePower: types.OnePower{
		Type:     "domain",
		Domain:   "DOME",
		Function: "ProductOffering",
		Action:   []string{"Update"},
	},
	ProductDeletePower: types.OnePower{
		Type:     "domain",
		Domain:   "DOME",
		Function: "ProductOffering",
		Action:   []string{"Delete"},
	},

	PolicyFileName: "auth_policies.star",
	VerifierServer: "https://verifier.dome-marketplace.eu",
	Dbname:         "data/isbetmf.db",
	ClonePeriod:    DefaultClonePeriod,
	Features: Features{
		OfferingLaunchOnlyByAdmin: true,
	},
}

var isbedevConfig = &Config{
	Environment:  ISBE_DEV,
	ProxyEnabled: false,

	ServerOperatorOrganizationIdentifier: "VATES-G87936159",
	ServerOperatorDid:                    "did:elsi:VATES-G87936159",
	ServerOperatorName:                   "Alastria",
	ServerOperatorCountry:                "ES",

	LEARPower: types.OnePower{
		Type:     "domain",
		Domain:   "DOME",
		Function: "Onboarding",
		Action:   []string{"execute"},
	},
	ProductCreatePower: types.OnePower{
		Type:     "domain",
		Domain:   "DOME",
		Function: "ProductOffering",
		Action:   []string{"Create"},
	},
	ProductUpdatePower: types.OnePower{
		Type:     "domain",
		Domain:   "DOME",
		Function: "ProductOffering",
		Action:   []string{"Update"},
	},
	ProductDeletePower: types.OnePower{
		Type:     "domain",
		Domain:   "DOME",
		Function: "ProductOffering",
		Action:   []string{"Delete"},
	},

	PolicyFileName: "auth_policies.star",
	VerifierServer: "https://verifier.dome-marketplace.eu",
	Dbname:         "data/isbetmf.db",
	ClonePeriod:    DefaultClonePeriod,
	Features: Features{
		OfferingLaunchOnlyByAdmin: true,
	},
}

var domeproConfig = &Config{
	Environment:  DOME_PRO,
	ProxyEnabled: true,

	LEARPower: types.OnePower{
		Type:     "domain",
		Domain:   "DOME",
		Function: "Onboarding",
		Action:   []string{"execute"},
	},
	ProductCreatePower: types.OnePower{
		Type:     "domain",
		Domain:   "DOME",
		Function: "ProductOffering",
		Action:   []string{"Create"},
	},
	ProductUpdatePower: types.OnePower{
		Type:     "domain",
		Domain:   "DOME",
		Function: "ProductOffering",
		Action:   []string{"Update"},
	},
	ProductDeletePower: types.OnePower{
		Type:     "domain",
		Domain:   "DOME",
		Function: "ProductOffering",
		Action:   []string{"Delete"},
	},

	PolicyFileName:  "auth_policies.star",
	RemoteTMFServer: "https://tmf.dome-marketplace.eu",
	VerifierServer:  "https://verifier.dome-marketplace.eu",
	Dbname:          "data/tmf.dome.pro.db",
	ClonePeriod:     DefaultClonePeriod,
}

var domepreConfig = &Config{
	Environment:  DOME_PRE,
	ProxyEnabled: true,

	LEARPower: types.OnePower{
		Type:     "domain",
		Domain:   "DOME",
		Function: "Onboarding",
		Action:   []string{"execute"},
	},
	ProductCreatePower: types.OnePower{
		Type:     "domain",
		Domain:   "DOME",
		Function: "ProductOffering",
		Action:   []string{"Create"},
	},
	ProductUpdatePower: types.OnePower{
		Type:     "domain",
		Domain:   "DOME",
		Function: "ProductOffering",
		Action:   []string{"Update"},
	},
	ProductDeletePower: types.OnePower{
		Type:     "domain",
		Domain:   "DOME",
		Function: "ProductOffering",
		Action:   []string{"Delete"},
	},

	PolicyFileName:  "auth_policies.star",
	RemoteTMFServer: "https://tmf.dome-marketplace-dev2.org",
	VerifierServer:  "https://verifier.dome-marketplace-dev2.org",
	Dbname:          "data/tmf.dome.dev2.db",
	ClonePeriod:     DefaultClonePeriod,
}

var lclConfig = &Config{
	Environment:  DOME_LCL,
	ProxyEnabled: true,

	LEARPower: types.OnePower{
		Type:     "domain",
		Domain:   "DOME",
		Function: "Onboarding",
		Action:   []string{"execute"},
	},
	ProductCreatePower: types.OnePower{
		Type:     "domain",
		Domain:   "DOME",
		Function: "ProductOffering",
		Action:   []string{"Create"},
	},
	ProductUpdatePower: types.OnePower{
		Type:     "domain",
		Domain:   "DOME",
		Function: "ProductOffering",
		Action:   []string{"Update"},
	},
	ProductDeletePower: types.OnePower{
		Type:     "domain",
		Domain:   "DOME",
		Function: "ProductOffering",
		Action:   []string{"Delete"},
	},

	PolicyFileName:  "auth_policies.star",
	RemoteTMFServer: "https://tmf.dome-marketplace-lcl.org",
	VerifierServer:  "https://verifier.dome-marketplace-lcl.org",
	Dbname:          "data/tmf.dome.lcl.db",
	ClonePeriod:     DefaultClonePeriod,
}
