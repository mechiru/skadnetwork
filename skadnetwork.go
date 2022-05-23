package skadnetwork

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Combine the values into a UTF-8 string with an invisible separator ('\u2063') between them.
// Version: 1.0, 2.0, 2.1, 2.2, 3.0
const separator = string('\u2063')

// Apple's public keys for postback:
var (
	// Apple's NIST P-192 public key that you use to verify postback version 1.0:
	// pubV1 = "MEkwEwYHKoZIzj0CAQYIKoZIzj0DAQEDMgAEMyHD625uvsmGq4C43cQ9BnfN2xslVT5V1nOmAMP6qaRRUll3PB1JYmgSm+62sosG"

	// Apple's NIST P-192 public key that you use to verify postback version 2.0:
	// pubV2 = "MEkwEwYHKoZIzj0CAQYIKoZIzj0DAQEDMgAEMyHD625uvsmGq4C43cQ9BnfN2xslVT5V1nOmAMP6qaRRUll3PB1JYmgSm"

	// Apple's NIST P-256 public key that you use to verify postback versions 2.1 or later:
	pubV3, _ = parseECDSAPublicKey("MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEWdp8GPcGqmhgzEFj9Z2nSpQVddayaPe4FMzqM9wib1+aHaaIzoHoLN9zW4K8y4SPykE3YVK3sVqW6Af0lfx3gg==")
)

// https://developer.apple.com/documentation/storekit/skadnetwork/signing_and_providing_ads
type FidelityType int

const (
	// App Store product page, rendered by StoreKit All SKAdNetwork versions.
	SKRenderedAds FidelityType = 1
	// Custom, provided by ad network SKAdNetwork version 2.2 and later.
	ViewThroughAds FidelityType = 0
)

func (f FidelityType) String() string {
	return strconv.FormatInt(int64(f), 10)
}

// https://developer.apple.com/documentation/storekit/skadnetwork/generating_the_signature_to_validate_storekit-rendered_ads
type Params struct {
	// Version 2.0 and later.
	// The SKAdNetwork API version number.
	Version string `json:"version,omitempty"`
	// Version 1.0 and later.
	// Your ad network identifier you registered with Apple.
	AdNetworkID string `json:"ad-network-id"`
	// Version 1.0 and later.
	// A campaign number you provide.
	CampaignID int `json:"campaign-id"`
	// Version 1.0 and later.
	// The App Store ID of the product to advertise.
	ItunesItemID int64 `json:"itunes-item-id"`
	// Version 1.0 and later.
	// A unique UUID value that you provide for each ad impression.
	// You must lowercase the string representation of the nonce value in the signature.
	Nonce uuid.UUID `json:"nonce"`
	// Version 2.0 and later.
	// The App Store ID of the app that displays the ad.
	// During testing, use an ID of 0 if you’re using a development-signed build and not an app from App Store.
	SourceAppStoreID int64 `json:"source-app-store-id,omitempty"`
	// Version 2.2 and later.
	// A value of 0 indicates a view-through ad presentation; a value of 1 indicates a StoreKit-rendered ad.
	FidelityType FidelityType `json:"fidelity-type,omitempty"`
	// Version 1.0 and later.
	// A timestamp you generate near the time of the ad impression.
	Timestamp time.Time `json:"timestamp"`
}

func (p Params) toItems() []string {
	items := make([]string, 0, 8)
	items = append(items,
		p.Version,
		p.AdNetworkID,
		strconv.Itoa(p.CampaignID),
		strconv.FormatInt(p.ItunesItemID, 10),
		p.Nonce.String(),
		strconv.FormatInt(p.SourceAppStoreID, 10),
	)
	switch p.Version {
	case "2.2", "3.0":
		items = append(items, p.FidelityType.String())
	}
	// Represented as UNIX time in milliseconds.
	items = append(items, strconv.FormatInt(p.Timestamp.UnixMilli(), 10))
	return items
}

type Postback struct {
	// Version 2.0 and later.
	// Set this key to version number "3.0", "2.2", “2.1", or "2.0".
	// For version availability, see SKAdNetwork Release Notes.
	Version string `json:"version,omitempty"`
	// Version 1.0 and later.
	// Your ad network identifier, that matches the value you provided
	// for SKStoreProductParameterAdNetworkIdentifier or adNetworkIdentifier.
	AdNetworkID string `json:"ad-network-id"`
	// Version 1.0 and later.
	// A unique value for this validation; use to deduplicate install validation messages.
	TransactionID string `json:"transaction-id"`
	// Version 1.0 and later.
	// The campaign ID you provided when displaying the ad,
	// that matches SKStoreProductParameterAdNetworkCampaignIdentifier or adCampaignIdentifier.
	CampaignID int `json:"campaign-id"`
	// Version 1.0 and later.
	// The item identifier of the advertised product.
	AppID int64 `json:"app-id"`
	// Version 2.0 and later.
	// Apple’s attribution signature, that you verify.
	AttributionSignature string `json:"attribution-signature,omitempty"`
	// Version 2.0 and later.
	// A Boolean flag that indicates that the customer redownloaded
	// and reinstalled the app when the value is true.
	Redownload *bool `json:"redownload,omitempty"`
	// Version 2.0 and later.
	// The item identifier of the app that displayed the ad.
	// During testing, if you’re using a development-signed build to display the ads
	// and not an app from App Store, use 0 as the item identifier.
	// Note: The source-app-id only appears in the postback if providing the parameter meets Apple’s privacy threshold.
	SourceAppID *int64 `json:"source-app-id,omitempty"`
	// Version 2.2 and later.
	// A value of 0 indicates a view-through ad presentation; a value of 1 indicates a StoreKit-rendered ad.
	FidelityType *FidelityType `json:"fidelity-type,omitempty"`
	// Version 2.0 and later.
	// An unsigned 6-bit value that the installed app provided by calling updateConversionValue(_:).
	// Note: The conversion-value only appears in the postback if the installed app provides it,
	// and if providing the parameter meets Apple’s privacy threshold.
	ConversionValue *uint8 `json:"conversion-value,omitempty"`
	// Version 3.0 and later.
	// A Boolean value that’s true if the ad network won the attribution,
	// and false if the postback represents a qualifying ad impression that didn’t win the attribution.
	DidWin *bool `json:"did-win,omitempty"`
}

// For 2.1:
// https://developer.apple.com/documentation/storekit/skadnetwork/verifying_an_install-validation_postback/combining_parameters_for_previous_postback_versions#3626226
//
// For 2.2:
// https://developer.apple.com/documentation/storekit/skadnetwork/verifying_an_install-validation_postback/combining_parameters_for_previous_postback_versions#3761660
//
// For 3.0:
// https://developer.apple.com/documentation/storekit/skadnetwork/verifying_an_install-validation_postback#2960703
func (p Postback) toItems() []string {
	ret := make([]string, 0, 9)
	ret = append(ret,
		p.Version,
		p.AdNetworkID,
		strconv.Itoa(p.CampaignID),
		strconv.FormatInt(p.AppID, 10),
		p.TransactionID,
		strconv.FormatBool(*p.Redownload),
	)
	if p.SourceAppID != nil {
		ret = append(ret, strconv.FormatInt(*p.SourceAppID, 10))
	}
	switch p.Version {
	case "2.2":
		ret = append(ret, p.FidelityType.String())
	case "3.0":
		ret = append(ret, p.FidelityType.String(), strconv.FormatBool(*p.DidWin))
	}
	return ret
}

func (p Postback) verify() (bool, error) {
	switch p.Version {
	case "2.1", "2.2", "3.0":
		return verify(pubV3, p.toItems(), p.AttributionSignature)
	default:
		return false, fmt.Errorf("skadnetwork: unsupported version error: %s", p.Version)
	}
}

type Signer struct {
	key *ecdsa.PrivateKey
}

func NewSigner(pem string) (*Signer, error) {
	key, err := decodePEM(pem)
	if err != nil {
		return nil, fmt.Errorf("skadnetwork: pem data decode error: %w", err)
	}
	return &Signer{key: key}, nil
}

func (s *Signer) sign(msg string) (string, error) {
	hash := hash(msg)
	der, err := ecdsa.SignASN1(rand.Reader, s.key, hash)
	if err != nil {
		return "", fmt.Errorf("skadnetwork: sign message error: %w", err)
	}
	return base64.StdEncoding.EncodeToString(der), nil
}

func (s *Signer) Sign(p *Params) (string, error) {
	msg := strings.Join(p.toItems(), separator)
	return s.sign(msg)
}

func (s *Signer) Verify(p *Params, sig string) (bool, error) {
	return verify(&s.key.PublicKey, p.toItems(), sig)
}

// https://developer.apple.com/documentation/storekit/skadnetwork/verifying_an_install-validation_postback#3599761
func Verify(p Postback) (bool, error) {
	return p.verify()
}

func parseECDSAPublicKey(s string) (*ecdsa.PublicKey, error) {
	der, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("skadnetwork: public key block decode error: %w", err)
	}

	pub, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return nil, fmt.Errorf("skadnetwork: parse public key error: %w", err)
	}
	return pub.(*ecdsa.PublicKey), nil
}

func hash(msg string) []byte {
	hash := sha256.Sum256([]byte(msg))
	return hash[:]
}

func verify(key *ecdsa.PublicKey, items []string, sig string) (bool, error) {
	der, err := base64.StdEncoding.DecodeString(sig)
	if err != nil {
		return false, fmt.Errorf("skadnetwork: signature decode error: %s", sig)
	}

	msg := strings.Join(items, separator)
	hash := hash(msg)
	return ecdsa.VerifyASN1(key, hash, der), nil
}

func decodePEM(data string) (*ecdsa.PrivateKey, error) {
	// https://stackoverflow.com/questions/21322182/how-to-store-ecdsa-private-key-in-go
	block, rest := pem.Decode([]byte(data))
	if block == nil {
		return nil, errors.New("skadnetwork: can not found data block")
	}

	if len(rest) != 0 {
		return nil, errors.New("skadnetwork: only 2 blocks are allowed for pem data")
	}

	if block.Type != "EC PRIVATE KEY" {
		return nil, fmt.Errorf("skadnetwork: unexpected block type detected: %s", block.Type)
	}
	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("skadnetwork: block parse error of pem data: %w", err)
	}
	return key, nil
}
