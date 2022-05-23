# skadnetwork

[![ci](https://github.com/mechiru/skadnetwork/workflows/ci/badge.svg)](https://github.com/mechiru/skadnetwork/actions?query=workflow:ci)

This library provides an implementation of [skadnetwork](https://developer.apple.com/documentation/storekit/skadnetwork).


## Note
Parameter signing and postback validation only support version **2.1 and above**.


## Example

Sign the parameter data:
```go
s, _ := skadnetwork.NewSigner(
    "-----BEGIN EC PRIVATE KEY-----\n...\n-----END EC PRIVATE KEY-----"
)
nonce := uuid.MustParse("68483ef6-0ada-40df-ab6b-3d19a66330fa")
timestamp, _ := time.Parse(time.RFC3339, "2022-05-06T10:00:00Z")

// For version 2.1
sig, _ := s.Sign(&skadnetwork.Params{
    Version:          "2.1",
    AdNetworkID:      "example123.skadnetwork",
    CampaignID:       42,
    ItunesItemID:     525463029,
    Nonce:            nonce,
    SourceAppStoreID: 1234567891,
    Timestamp:        timestamp,
})

// For version 2.2
sig, _ := s.Sign(&skadnetwork.Params{
    Version:          "2.2",
    AdNetworkID:      "example123.skadnetwork",
    CampaignID:       42,
    ItunesItemID:     525463029,
    Nonce:            nonce,
    SourceAppStoreID: 1234567891,
    FidelityType: skadnetwork.SKRenderedAds,
    Timestamp:        timestamp,
})

// For version 3.0
sig, _ := s.Sign(&skadnetwork.Params{
    Version:          "3.0",
    AdNetworkID:      "example123.skadnetwork",
    CampaignID:       42,
    ItunesItemID:     525463029,
    Nonce:            nonce,
    SourceAppStoreID: 1234567891,
    FidelityType: skadnetwork.SKRenderedAds,
    Timestamp:        timestamp,
})
```

Verify the signed parameter data:
```go
ok, _ := s.Verify(&skadnetwork.Params{ ... }, sig)
```

Verify the Apple's postback data:
```go
ok, _ := skadnetwork.Verify(&skadnetwork.Postback{ ... })
```


## How to generate your public-private key pair

Please see [Apple's documentation](https://developer.apple.com/documentation/storekit/skadnetwork/registering_an_ad_network#3657881).

```bash
# Generate private key
openssl ecparam -name prime256v1 -genkey -noout -out companyname_skadnetwork_private_key.pem

# Generate pubilc key
openssl ec -in companyname_skadnetwork_private_key.pem -pubout -out companyname_skadnetwork_public_key.pem
```


## License
Licensed under the [MIT license](./LICENSE).
