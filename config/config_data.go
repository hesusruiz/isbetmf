package config

import "github.com/hesusruiz/isbetmf/types"

// As this PDP is designed for DOME and ISBE environments, many config data items are configured in code.
// This avoids many configuration errors and simplifies deployment, at the expense of some flexibility.
// However, this flexibility is not really needed in practice, as the DOME environments are well defined and stable.
// Minimizing errors is here much more important than the ease to configure these parameters.

// ******************************************************************
// ISBE configurations
// ******************************************************************

var isbedevConfig = &Config{
	Environment:  ISBE_DEV,
	ProxyEnabled: false,

	// The operator is Alastria
	ServerOperatorOrganizationIdentifier: "VATES-G87936159",
	ServerOperatorDid:                    "did:elsi:VATES-G87936159",
	ServerOperatorName:                   "Alastria",
	ServerOperatorCountry:                "ES",

	LEARPower: types.OnePower{
		Type:     "organization",
		Domain:   "ISBE",
		Function: "Onboarding",
		Action:   []string{"execute"},
	},
	ProductCreatePower: types.OnePower{
		Type:     "organization",
		Domain:   "ISBE",
		Function: "ProductOffering",
		Action:   []string{"Create"},
	},
	ProductUpdatePower: types.OnePower{
		Type:     "organization",
		Domain:   "ISBE",
		Function: "ProductOffering",
		Action:   []string{"Update"},
	},
	ProductDeletePower: types.OnePower{
		Type:     "organization",
		Domain:   "ISBE",
		Function: "ProductOffering",
		Action:   []string{"Delete"},
	},

	PolicyFileName: "auth_policies.star",
	VerifierServer: "https://idp.dev.cloud-w.envs.redisbe.com/auth/realms/dev-isbe",

	Dbname:      "data/isbetmf.db",
	ClonePeriod: DefaultClonePeriod,
	Features: Features{
		OfferingLaunchOnlyByAdmin: true,
		GenerateIDOnCreate:        true,
		AllowIDInBody:             true,
		VerifyJWTSignature:        true,
	},
}

var isbepreConfig = &Config{
	Environment:  ISBE_PRE,
	ProxyEnabled: false,

	// The operator is Alastria
	ServerOperatorOrganizationIdentifier: "VATES-G87936159",
	ServerOperatorDid:                    "did:elsi:VATES-G87936159",
	ServerOperatorName:                   "Alastria",
	ServerOperatorCountry:                "ES",

	LEARPower: types.OnePower{
		Type:     "organization",
		Domain:   "ISBE",
		Function: "Onboarding",
		Action:   []string{"execute"},
	},
	ProductCreatePower: types.OnePower{
		Type:     "organization",
		Domain:   "ISBE",
		Function: "ProductOffering",
		Action:   []string{"Create"},
	},
	ProductUpdatePower: types.OnePower{
		Type:     "organization",
		Domain:   "ISBE",
		Function: "ProductOffering",
		Action:   []string{"Update"},
	},
	ProductDeletePower: types.OnePower{
		Type:     "organization",
		Domain:   "ISBE",
		Function: "ProductOffering",
		Action:   []string{"Delete"},
	},

	PolicyFileName: "auth_policies.star",
	VerifierServer: "https://idp.pre.portal.redisbe.com/auth/realms/pre-isbe",

	Dbname:      "data/isbetmf.db",
	ClonePeriod: DefaultClonePeriod,
	Features: Features{
		OfferingLaunchOnlyByAdmin: true,
		GenerateIDOnCreate:        true,
		AllowIDInBody:             true,
		VerifyJWTSignature:        true,
	},
}

var isbeproConfig = &Config{
	Environment:  ISBE_PRO,
	ProxyEnabled: false,

	// The operator is Alastria
	ServerOperatorOrganizationIdentifier: "VATES-G87936159",
	ServerOperatorDid:                    "did:elsi:VATES-G87936159",
	ServerOperatorName:                   "Alastria",
	ServerOperatorCountry:                "ES",

	LEARPower: types.OnePower{
		Type:     "organization",
		Domain:   "ISBE",
		Function: "Onboarding",
		Action:   []string{"execute"},
	},
	ProductCreatePower: types.OnePower{
		Type:     "organization",
		Domain:   "ISBE",
		Function: "ProductOffering",
		Action:   []string{"Create"},
	},
	ProductUpdatePower: types.OnePower{
		Type:     "organization",
		Domain:   "ISBE",
		Function: "ProductOffering",
		Action:   []string{"Update"},
	},
	ProductDeletePower: types.OnePower{
		Type:     "organization",
		Domain:   "ISBE",
		Function: "ProductOffering",
		Action:   []string{"Delete"},
	},

	PolicyFileName: "auth_policies.star",
	VerifierServer: "https://idp.portal.redisbe.com/auth/realms/pro-isbe",

	Dbname:      "data/isbetmf.db",
	ClonePeriod: DefaultClonePeriod,
	Features: Features{
		OfferingLaunchOnlyByAdmin: true,
		GenerateIDOnCreate:        true,
		AllowIDInBody:             true,
		VerifyJWTSignature:        true,
	},
}

// ******************************************************************
// DOME configurations
// ******************************************************************

// The DOME SBX environment is used for development and testing, so it is called DOME_DEV
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
	Features: Features{
		OfferingLaunchOnlyByAdmin: false,
		GenerateIDOnCreate:        false,
		AllowIDInBody:             false,
	},
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
	Features: Features{
		OfferingLaunchOnlyByAdmin: false,
		GenerateIDOnCreate:        false,
		AllowIDInBody:             false,
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
	Features: Features{
		OfferingLaunchOnlyByAdmin: false,
		GenerateIDOnCreate:        false,
		AllowIDInBody:             false,
	},
}

// ******************************************************************
// Local developmentconfiguration
// ******************************************************************

var lclConfig = &Config{
	Environment:  LOCAL,
	ProxyEnabled: false,

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
	VerifierServer: "https://verifier.dome-marketplace-sbx.org",
	Dbname:         "data/tmf.lcl.db",
	ClonePeriod:    DefaultClonePeriod,
	Features: Features{
		OfferingLaunchOnlyByAdmin: false,
		GenerateIDOnCreate:        true,
		AllowIDInBody:             false,
	},
}
