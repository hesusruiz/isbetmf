# TMForum Object Validation Report

**Generated:** 2025-09-04 12:00:48 UTC

**Configuration:**
- Base URL: `https://tmf.dome-marketplace-sbx.org`
- Object Types: customerBillOnDemand, individual, productOffering, resourceCandidate, resourceOrder, cancelProductOrder, productOfferingPrice, resourceCatalog, resourceFunction, serviceSpecification, billingAccount, cancelResourceOrder, customerBill, product, productSpecification, scale, agreement, appliedCustomerBillingRate, customer, organization, partyAccount, partyRole, productOrder, serviceCandidate, catalog, category, resourceCategory, serviceCategory, settlementAccount, billPresentationMedia, migrate, quote, service, serviceCatalog, usage, financialAccount, resourceSpecification, billFormat, cancelServiceOrder, heal, monitor, resource, serviceOrder, usageSpecification, agreementSpecification, billingCycleSpecification
- Timeout: 30 seconds
- Validate Required Fields: true
- Validate Related Party: true

## Summary Statistics

| Metric | Value |
|--------|-------|
| Total Objects | 199 |
| Valid Objects | 0 |
| Invalid Objects | 199 |
| Total Errors | 450 |
| Total Warnings | 0 |
| Processing Time | 35.571µs |

## Statistics by Object Type

| Object Type | Count | Valid | Invalid | Errors | Warnings |
|-------------|-------|-------|---------|--------|----------|
| appliedCustomerBillingRate | 6 | 0 | 6 | 27 | 0 |
| billingAccount | 6 | 0 | 6 | 36 | 0 |
| catalog | 14 | 0 | 14 | 53 | 0 |
| category | 78 | 0 | 78 | 78 | 0 |
| individual | 33 | 0 | 33 | 66 | 0 |
| organization | 13 | 0 | 13 | 26 | 0 |
| product | 1 | 0 | 1 | 6 | 0 |
| productOffering | 13 | 0 | 13 | 13 | 0 |
| productOfferingPrice | 1 | 0 | 1 | 2 | 0 |
| productOrder | 8 | 0 | 8 | 56 | 0 |
| productSpecification | 9 | 0 | 9 | 18 | 0 |
| resourceSpecification | 4 | 0 | 4 | 12 | 0 |
| serviceSpecification | 7 | 0 | 7 | 21 | 0 |
| usageSpecification | 6 | 0 | 6 | 36 | 0 |

## Error Summary

| Error Code | Count |
|-------------|-------|
| MISSING_PARTY_NAME | 10 |
| MISSING_PARTY_REFERRED_TYPE | 9 |
| MISSING_RELATED_PARTY | 20 |
| MISSING_REQUIRED_FIELD | 263 |
| MISSING_REQUIRED_ROLE | 148 |

## Detailed Validation Results

### appliedCustomerBillingRate Objects

#### Object: urn:ngsi-ld:applied-customer-billing-rate:7ea1db4a-b6b6-489f-ad68-1e5d0644dcb6

- **Type:** appliedCustomerBillingRate
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Related party information is required but missing (Code: MISSING_RELATED_PARTY)

#### Object: urn:ngsi-ld:applied-customer-billing-rate:a886304d-d699-4adf-b93e-dcdcd54474f1

- **Type:** appliedCustomerBillingRate
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyer' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyeroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:applied-customer-billing-rate:c25e45c6-8116-44d8-8101-651fddd379e2

- **Type:** appliedCustomerBillingRate
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyer' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyeroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:applied-customer-billing-rate:144d5a4a-6b0b-4308-beec-f15cddce3cbd

- **Type:** appliedCustomerBillingRate
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyer' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyeroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:applied-customer-billing-rate:0bd7ffd4-4a34-4e9c-8dc8-dfd60818bf65

- **Type:** appliedCustomerBillingRate
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Related party information is required but missing (Code: MISSING_RELATED_PARTY)

#### Object: urn:ngsi-ld:applied-customer-billing-rate:9b2925a3-775b-4996-8f7d-7d70ca15e367

- **Type:** appliedCustomerBillingRate
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Related party information is required but missing (Code: MISSING_RELATED_PARTY)

### billingAccount Objects

#### Object: urn:ngsi-ld:billing-account:3b49919b-ab08-4969-ab53-cfd06cc21206

- **Type:** billingAccount
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyer' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyeroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:billing-account:d2c224df-b007-4524-9029-7e7e1b021d35

- **Type:** billingAccount
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyer' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyeroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:billing-account:c84d03ff-fc74-435c-a54c-fed6e95ff80a

- **Type:** billingAccount
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyer' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyeroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:billing-account:f98a4654-612b-474a-875b-107243e814bb

- **Type:** billingAccount
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyer' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyeroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:billing-account:5a7a0a3f-61f8-4b62-bdf9-5212cbfc129c

- **Type:** billingAccount
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyer' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyeroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:billing-account:f727fb85-51f5-4dca-822f-4f4bb2775549

- **Type:** billingAccount
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyer' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyeroperator' is missing (Code: MISSING_REQUIRED_ROLE)

### catalog Objects

#### Object: urn:ngsi-ld:catalog:1d535a9b-212c-4e8b-aaff-8a412e61dd0d

- **Type:** catalog
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:catalog:e40feb38-04b6-485b-9edc-789704d3cd85

- **Type:** catalog
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:catalog:03159bce-35ca-4938-bc2b-b8239e1008ca

- **Type:** catalog
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:catalog:c016bd48-8d84-48c9-82f1-8cb1d0c1cddd

- **Type:** catalog
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:catalog:4a86c5a6-610f-47b0-8c0f-6bbf472409da

- **Type:** catalog
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:catalog:44e13ee8-c1ee-4c7a-a0f0-a308cd73a0e3

- **Type:** catalog
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:catalog:ec45d9d1-50a1-445d-848e-09f50ca7862e

- **Type:** catalog
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:catalog:ad73d40f-11b6-4275-ab8a-6d1983e7b3ee

- **Type:** catalog
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:catalog:fb58a661-0341-4b61-9dac-610e228ad6bd

- **Type:** catalog
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:catalog:2df6d831-26d8-430f-9ac7-79cf447c47fe

- **Type:** catalog
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Related party information is required but missing (Code: MISSING_RELATED_PARTY)

#### Object: urn:ngsi-ld:catalog:721d9e67-0a46-4126-a12e-8f91670ceaf7

- **Type:** catalog
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:catalog:dfd8bf69-cb7e-4d0b-a3de-f407b2849580

- **Type:** catalog
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:catalog:b2a1b435-d234-4f58-a23f-92c6ba5798b5

- **Type:** catalog
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Related party information is required but missing (Code: MISSING_RELATED_PARTY)

#### Object: urn:ngsi-ld:catalog:134206dc-e658-4f12-a59b-ad2e17f4ede4

- **Type:** catalog
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Related party information is required but missing (Code: MISSING_RELATED_PARTY)

### category Objects

#### Object: urn:ngsi-ld:category:8d222bda-159e-4957-8b76-0fb06b4449dd

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:d6aa17c8-a0d4-4312-9ab1-ae78a0aace1d

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:73dd1ca0-defd-4642-bbec-f44d52273973

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:3867db98-1eb3-4bff-8a23-715082f404ff

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:d64a7aff-f49c-4cbf-84d8-4b883e59e392

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:ef149b50-08ea-4ee2-bf41-a90b04a1ce79

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:55b2cb9f-e13b-499e-bf76-fd1eb56d2bde

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:57158a09-c62b-4160-a49e-2295ca068682

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:3cad4d15-bbe6-44d6-a82a-a6478b8e53ce

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:35182b9c-be04-4a40-ae67-f9433f4d36e1

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:b332fe31-ec4d-4883-b051-e663a3d40830

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:3ae62a2d-a194-4250-9b3f-8c9349ef4799

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:bb19f802-c0dd-4686-9049-cbd4665e7a04

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:44370eaa-8828-4e1d-b9be-03b64eb11309

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:44e33c19-dd19-4ee9-a816-9dea2d34f02c

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:a7779dea-c0c0-47af-b078-3098c309cc23

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:840bf629-6e64-4e49-94c0-740d7b312dd4

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:7c796f0f-8e89-4578-bd6f-5d6c56ec912d

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:d154f150-2980-4b5d-9aef-2fdc123dc4d9

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:dc159e2d-d0a1-49e2-a2ec-b7c4c97e4dd7

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:2b015a9f-50e6-453f-a251-e8770e2260bb

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:7aa7ae47-4c3d-447e-848b-60b3f015cb49

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:9bfdb312-ad44-41c5-9c5a-4d58ad861f63

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:15f52622-9097-4fba-a66b-435fb651a393

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:ed12cd4a-3a23-46fe-8135-ea5bdfbb2b4b

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:bd99e5e4-5143-470e-9c05-d7839dfc71bc

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:b51bdeb5-0cf0-41a3-b0af-3ef48a88c988

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:17c3aafe-2c2d-49be-8dce-dd4981e22e28

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:ffeae64f-edc2-4754-906e-a4d35d10c806

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:91abfd48-4870-4af1-b64f-69cb144d1dc6

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:a14a4067-a5c9-4122-b2a7-28ee2dd85036

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:08892d05-2e7a-41f9-8e8b-2dcc28f86b8a

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:abe28b41-cc3a-4624-9c23-2ed3e021dbb5

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:96778468-41b5-4245-842e-f3c481a240b0

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:8337bace-8228-4925-a75a-e2654f0bac41

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:c053c314-6696-49bd-a10b-052e6f3dbddb

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:94bd8c25-4183-4124-a6ed-0516d4ced4a6

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:630359eb-7d66-4eb2-bf07-2590841049a0

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:af87fc59-aafb-4301-8ba9-3594558a9eff

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:9c02c10e-1cdf-4de3-a99e-f0b689881d2b

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:8795e633-8a64-438f-8cad-a2d5175cc9cb

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:47eeb384-4400-49c1-b0e5-718ef17860ba

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:9a7f4ac4-e481-4303-98fc-b4f9911f6529

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:03e76d62-a8c6-4571-ad7a-c31e754d3d3f

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:6ff41b67-bf97-4772-86ef-52d1d4998a3e

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:228fb947-48df-4281-b95d-c675bdc31b7a

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:419a8fac-d55a-4ae7-867f-2a3a338844cb

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:061ba18d-4263-42f3-9d6e-17d90ad25832

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:0b5cb3a5-b08e-4fca-8fb4-f99afaf71a2a

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:b1f8c6b7-e73e-415a-9f28-51c01adc36c1

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:55ff6485-3bcf-4009-815a-fca65727c64d

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:cfb68599-fde0-457e-bd8d-f6f6db20025f

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:13837a4a-4eee-4879-be8b-da8bc88f8086

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:10ba35dd-4906-4434-adac-9bb9268e1127

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:ef80f5f1-c480-48f8-b96b-b893efec3e4d

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:b4998d10-d9c2-46ad-a08d-63df7cd8e911

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:3beea627-9691-4ed8-9bbf-7c6fe06e278a

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:3ebbe833-9f85-42d2-b76d-7429d156eba6

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:818597c8-6de7-4961-8ce3-61fb30d7fbb8

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:a456f7a7-cbb0-4b71-9c25-f8cb46555016

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:2cc758fa-80b3-4704-9c38-26918c25b535

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:9c12beda-8e3b-4b90-ba96-191e0129a9d4

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:6df17fbc-465d-4e0e-93af-ecc0ccdf34b1

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:f0c23dff-8f09-474a-8306-607ca835d8ce

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:c9f5f9b8-2991-4cb5-aad9-eede449c88e8

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:4c9b41bb-2eeb-4229-8442-b0e8d8127271

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:917fa857-f354-49f1-84f3-38540e07e434

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:ac6eefc6-1e58-432f-a9b9-c8349c3a3d68

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:465a29be-b24c-448e-80ed-c9a1e5aeb3da

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:63ed0da1-bc76-4e8d-af4d-f551d8148394

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:1d731f47-e2ce-4f2a-899b-f7f1c449cd98

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:fcf925b0-bb66-4860-8c2a-db17861fe105

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:8ec9aa71-afa0-4ae4-93d2-61eef79fe9c6

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:2fffafeb-64ca-4d72-9e4c-ed7d2973b3a9

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:b664a377-e12c-465b-ae2c-0c5a53438e65

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:a6f7ec89-e2e3-429d-a16c-30b0a1ab0ff6

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:814efafb-b2be-4d7e-b83d-0d447d9f68ac

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:category:4b11ef5c-50a6-4a20-8817-8152b372cc95

- **Type:** category
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

### individual Objects

#### Object: urn:ngsi-ld:individual:97790bad-066f-41fa-bb5a-745d783e044e

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:individual:038e6cf4-8805-4905-a151-42012ddcdfbf

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:individual:c1046c38-844f-471f-873c-78b35c270c27

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:individual:afa0bf6f-0b1e-4d9c-8529-e57bd28740f8

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:individual:911f60f5-9a54-4d1c-ae99-46079d5c6143

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:individual:9506e609-cf28-4302-9592-7a49521d0412

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:individual:c5b57368-f3cc-46ed-838d-25153f471c5d

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:individual:d6cccbbb-3fae-413f-89b9-c3d41ea25362

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:individual:73ecc3c1-41b9-48a3-8c99-b29e6e02ad2d

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:individual:8f307e88-4208-4484-9a62-b06b9b4d4a29

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:individual:9e2e4859-c2f4-491c-a787-5f3aeec93ef6

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:individual:b8d107f5-5888-413c-a334-df87468398ad

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:individual:d3b2d0a8-d490-43df-8f00-f76b45cae5af

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:individual:68f857d5-4de3-4f22-8daf-0b46d9699a8d

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:individual:5c9d40b0-5628-4dec-9228-08fbee4bcce2

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:individual:46b73e37-a453-4b19-8df0-85356a52cd8a

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:individual:b266d7da-0852-4659-a003-820b391f7927

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:individual:98d935c4-5f25-42de-a000-844950f1b766

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:individual:0774f8c5-073e-4334-97b2-5b71167dfc2b

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:individual:a0e4422b-1464-42a5-ac76-2901329a5c8b

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:individual:b068358a-d4c0-4f20-a522-ac6286181d49

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:individual:c475f14f-97b3-45ec-8a34-4f22f4931f04

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:individual:f366c06b-38b2-4039-b66a-8dde9a88f1f2

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:individual:80bb5420-a8c4-47a6-aa73-8c96ccc9a525

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:individual:6aebe06a-54d2-4f17-87c4-940c1e22ef91

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:individual:bb26a135-8281-430c-b2d0-652b752b0c15

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:individual:5e7c0d3f-235e-4507-9875-a166fb41fb35

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:individual:551f1882-63c9-44fb-b001-d9c2648ba6bc

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:individual:16a87ab8-2d62-4298-97dd-6c96f2556397

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:individual:521e32a5-bdc3-4530-a878-644299792594

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:individual:2419897b-0da3-41f2-a34f-d92ccff3a9e2

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:individual:67c4985f-2ed1-43d2-a631-6261a66f7c8e

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:individual:6176fae2-ef4e-4bb5-aae6-f7d6e41efc3f

- **Type:** individual
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

### organization Objects

#### Object: urn:ngsi-ld:organization:00026ead-d16c-40ee-9218-48defc7ce749

- **Type:** organization
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:organization:a2f5ebea-49c9-4015-a9d6-56f2c566f6c9

- **Type:** organization
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:organization:d8dcc7a3-0774-4824-b552-d05d91986565

- **Type:** organization
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:organization:c2c39ab0-1f15-40f9-89ce-fc131953d33e

- **Type:** organization
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:organization:2bcbe859-e316-42f2-919c-f470cff9e235

- **Type:** organization
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:organization:eb6647da-84f2-4645-8d9f-c2905775b561

- **Type:** organization
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:organization:95fdc12e-6889-4f08-8ff8-296b10e8e781

- **Type:** organization
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:organization:e4fa0e9f-1779-49c0-9e8a-66a02bf1fe4e

- **Type:** organization
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:organization:4eb54a0c-a916-499a-bf5f-6bba76101e1e

- **Type:** organization
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:organization:fb47dd79-6497-4ec8-8456-e6e06d3c698b

- **Type:** organization
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:organization:44c33614-df98-459e-b710-064b1a7c6f65

- **Type:** organization
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:organization:6c53e937-212b-4e8b-997c-4d8695f789d1

- **Type:** organization
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

#### Object: urn:ngsi-ld:organization:338282a4-3c06-41d0-8c35-3fe5cecc38db

- **Type:** organization
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)

### product Objects

#### Object: urn:ngsi-ld:product:45eaa84b-122f-4e6d-9797-31ab1ab16134

- **Type:** product
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty seller.name: Related party name is missing (Code: MISSING_PARTY_NAME)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyer' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyeroperator' is missing (Code: MISSING_REQUIRED_ROLE)

### productOffering Objects

#### Object: urn:ngsi-ld:product-offering:005974a1-f327-47bd-96fc-2c263f2818c4

- **Type:** productOffering
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - relatedParty: Related party information is required but missing (Code: MISSING_RELATED_PARTY)

#### Object: urn:ngsi-ld:product-offering:4174966c-aca0-4787-88fe-96dd235fa2df

- **Type:** productOffering
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - relatedParty: Related party information is required but missing (Code: MISSING_RELATED_PARTY)

#### Object: urn:ngsi-ld:product-offering:17f86010-d4c4-4120-99f5-25d25f376bbb

- **Type:** productOffering
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - relatedParty: Related party information is required but missing (Code: MISSING_RELATED_PARTY)

#### Object: urn:ngsi-ld:product-offering:9bbc3d54-daae-414e-a63a-52fa06dfcd0f

- **Type:** productOffering
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - relatedParty: Related party information is required but missing (Code: MISSING_RELATED_PARTY)

#### Object: urn:ngsi-ld:product-offering:dcdd91b1-22ee-41e2-968b-24289c077e18

- **Type:** productOffering
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - relatedParty: Related party information is required but missing (Code: MISSING_RELATED_PARTY)

#### Object: urn:ngsi-ld:product-offering:7bbf2620-43fe-4afe-b2ff-cfe83f78484a

- **Type:** productOffering
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - relatedParty: Related party information is required but missing (Code: MISSING_RELATED_PARTY)

#### Object: urn:ngsi-ld:product-offering:1c9bdbba-e2bf-4b24-8586-9d01f8900cf1

- **Type:** productOffering
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - relatedParty: Related party information is required but missing (Code: MISSING_RELATED_PARTY)

#### Object: urn:ngsi-ld:product-offering:66b48359-1d47-42af-a974-1c40ee50f3dd

- **Type:** productOffering
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - relatedParty: Related party information is required but missing (Code: MISSING_RELATED_PARTY)

#### Object: urn:ngsi-ld:product-offering:54a9af6f-f353-48a5-b599-84a535f0bc74

- **Type:** productOffering
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - relatedParty: Related party information is required but missing (Code: MISSING_RELATED_PARTY)

#### Object: urn:ngsi-ld:product-offering:4699bdc1-592a-42b0-ad2d-93b0a06b044a

- **Type:** productOffering
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - relatedParty: Related party information is required but missing (Code: MISSING_RELATED_PARTY)

#### Object: urn:ngsi-ld:product-offering:038df757-e986-4822-97a9-9934c903340e

- **Type:** productOffering
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - relatedParty: Related party information is required but missing (Code: MISSING_RELATED_PARTY)

#### Object: urn:ngsi-ld:product-offering:2135b6c4-6de8-4ec9-9790-a3a322650a8a

- **Type:** productOffering
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - relatedParty: Related party information is required but missing (Code: MISSING_RELATED_PARTY)

#### Object: urn:ngsi-ld:product-offering:5a4bab89-9f7c-498b-bde8-f70324488b4f

- **Type:** productOffering
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - relatedParty: Related party information is required but missing (Code: MISSING_RELATED_PARTY)

### productOfferingPrice Objects

#### Object: urn:ngsi-ld:product-offering-price:91a5b7f3-afb1-427c-bed1-85332ee1448d

- **Type:** productOfferingPrice
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Related party information is required but missing (Code: MISSING_RELATED_PARTY)

### productOrder Objects

#### Object: urn:ngsi-ld:product-order:fdc85cc5-6e74-4e67-a9d0-1f3d7a79dce3

- **Type:** productOrder
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty seller.name: Related party name is missing (Code: MISSING_PARTY_NAME)
  - relatedParty seller.referredType: Related party referred type is missing (Code: MISSING_PARTY_REFERRED_TYPE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyer' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyeroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:product-order:4902aa01-8949-4a46-9be4-1cf98aa4a4f7

- **Type:** productOrder
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty seller.name: Related party name is missing (Code: MISSING_PARTY_NAME)
  - relatedParty seller.referredType: Related party referred type is missing (Code: MISSING_PARTY_REFERRED_TYPE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyer' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyeroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:product-order:117780df-ea0f-45b9-bcb3-1cbe5b17b1f9

- **Type:** productOrder
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty seller.name: Related party name is missing (Code: MISSING_PARTY_NAME)
  - relatedParty seller.referredType: Related party referred type is missing (Code: MISSING_PARTY_REFERRED_TYPE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyer' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyeroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:product-order:06350c8a-4238-4284-a9b4-f5c047b91292

- **Type:** productOrder
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty seller.name: Related party name is missing (Code: MISSING_PARTY_NAME)
  - relatedParty seller.referredType: Related party referred type is missing (Code: MISSING_PARTY_REFERRED_TYPE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyer' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyeroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:product-order:4225f20e-f0d5-4c73-a7e7-b50fc86c6cd7

- **Type:** productOrder
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty seller.name: Related party name is missing (Code: MISSING_PARTY_NAME)
  - relatedParty seller.referredType: Related party referred type is missing (Code: MISSING_PARTY_REFERRED_TYPE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyer' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyeroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:product-order:9dee5ca2-947e-4443-aca8-f82ec936d2b2

- **Type:** productOrder
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty seller.name: Related party name is missing (Code: MISSING_PARTY_NAME)
  - relatedParty seller.referredType: Related party referred type is missing (Code: MISSING_PARTY_REFERRED_TYPE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyer' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyeroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:product-order:bf166eed-bfb9-4632-bf9c-3d407d817a19

- **Type:** productOrder
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty seller.name: Related party name is missing (Code: MISSING_PARTY_NAME)
  - relatedParty seller.referredType: Related party referred type is missing (Code: MISSING_PARTY_REFERRED_TYPE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyer' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyeroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:product-order:fa5e91ba-6418-4354-b840-1f9e31ca1e74

- **Type:** productOrder
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty seller.name: Related party name is missing (Code: MISSING_PARTY_NAME)
  - relatedParty seller.referredType: Related party referred type is missing (Code: MISSING_PARTY_REFERRED_TYPE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyer' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyeroperator' is missing (Code: MISSING_REQUIRED_ROLE)

### productSpecification Objects

#### Object: urn:ngsi-ld:product-specification:bf72e349-03f0-4e88-965a-73815d8881b4

- **Type:** productSpecification
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:product-specification:55158c3d-ef8f-4f1c-b9d0-82dd3138b2ae

- **Type:** productSpecification
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:product-specification:5e2ced54-45f1-4687-b9f9-ee13cb318a66

- **Type:** productSpecification
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:product-specification:ccaec925-9fad-40c0-8015-5b9087f92aff

- **Type:** productSpecification
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:product-specification:ba4123f5-77f2-4b80-8c8d-13c01d3a6e72

- **Type:** productSpecification
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:product-specification:411816fd-3f8b-4e6f-8c35-02e8cada5ce9

- **Type:** productSpecification
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:product-specification:534697ce-c774-4cc2-a8f4-d4f2ae6cf4f9

- **Type:** productSpecification
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:product-specification:bfe7418d-2614-4e6a-920c-6d6711171f10

- **Type:** productSpecification
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:product-specification:10ebf2c1-fe79-4191-81a3-b58207307c5a

- **Type:** productSpecification
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)

### resourceSpecification Objects

#### Object: urn:ngsi-ld:resource-specification:9bd04c66-97a5-489a-ae93-7dddb4cd341b

- **Type:** resourceSpecification
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:resource-specification:f0153609-77d6-4464-9352-679f8d0e015f

- **Type:** resourceSpecification
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:resource-specification:e84e77c9-55e1-4c2f-953a-09dd52003f92

- **Type:** resourceSpecification
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:resource-specification:59cf3608-2879-4f38-a8f7-5baf61653c92

- **Type:** resourceSpecification
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)

### serviceSpecification Objects

#### Object: urn:ngsi-ld:service-specification:ab1f9684-e04f-4692-bb2d-20827f1bb759

- **Type:** serviceSpecification
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:service-specification:8f5f2d0c-9af4-47ad-a932-387455fc11df

- **Type:** serviceSpecification
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:service-specification:69a374ff-97c5-417b-8b31-2bd36798006b

- **Type:** serviceSpecification
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:service-specification:80b2ba93-5d5d-4753-a29f-b80114a01333

- **Type:** serviceSpecification
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:service-specification:5ff52024-ea17-4f17-aa39-0499f57fc7d1

- **Type:** serviceSpecification
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:service-specification:8a8741bb-c559-46a8-9ae8-d2f00c7504ca

- **Type:** serviceSpecification
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:service-specification:89d6bddb-7805-4042-a60a-030ef09ed816

- **Type:** serviceSpecification
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:47 UTC
- **Errors:**
  - relatedParty seller.name: Related party name is missing (Code: MISSING_PARTY_NAME)
  - relatedParty seller.referredType: Related party referred type is missing (Code: MISSING_PARTY_REFERRED_TYPE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)

### usageSpecification Objects

#### Object: urn:ngsi-ld:usageSpecification:21333496-4652-4bbf-be85-a897278d4ee9

- **Type:** usageSpecification
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyer' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyeroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:usageSpecification:24abecd5-bf1f-42e0-a34f-1657f39dffe1

- **Type:** usageSpecification
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyer' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyeroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:usageSpecification:156985fe-69c9-4f2f-a906-cbdc01d4d427

- **Type:** usageSpecification
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyer' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyeroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:usageSpecification:3aacf05a-838b-4160-a810-4447fc58695e

- **Type:** usageSpecification
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyer' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyeroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:usageSpecification:fd5da80e-fb29-45e8-8636-46b46dc2973a

- **Type:** usageSpecification
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyer' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyeroperator' is missing (Code: MISSING_REQUIRED_ROLE)

#### Object: urn:ngsi-ld:usageSpecification:e16ff3ad-0bf6-4b8c-95d9-0d2e69b4c12a

- **Type:** usageSpecification
- **Valid:** false
- **Timestamp:** 2025-09-04 12:00:48 UTC
- **Errors:**
  - lastUpdate: Required field 'lastUpdate' is missing (Code: MISSING_REQUIRED_FIELD)
  - version: Required field 'version' is missing (Code: MISSING_REQUIRED_FIELD)
  - relatedParty: Required related party role 'seller' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'selleroperator' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyer' is missing (Code: MISSING_REQUIRED_ROLE)
  - relatedParty: Required related party role 'buyeroperator' is missing (Code: MISSING_REQUIRED_ROLE)

---

*Report generated by TMForum Proxy Validator*
*Generated at: 2025-09-04 12:00:48 UTC*
