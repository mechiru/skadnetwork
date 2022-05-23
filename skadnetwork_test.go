package skadnetwork_test

import (
	"encoding/json"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/google/uuid"

	"github.com/mechiru/skadnetwork"
)

const (
	pem = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIPAYHdpbrKcTKi6qrRBB/TYN4w33jXAL0j9JMOqu5oIZoAoGCCqGSM49
AwEHoUQDQgAEBdF30K5pLjixuXnqiCNN/AgUK3DexfWqLzNOn2cZt0t9lMR8Y/Dl
MgSZN35Bv8gyUXt7xOK+hP8tDoOD2ir7bw==
-----END EC PRIVATE KEY-----
`

	v2_2 = `{
  "version" : "2.2",
  "ad-network-id" : "com.example",
  "campaign-id" : 42,
  "transaction-id" : "6aafb7a5-0170-41b5-bbe4-fe71dedf1e28",
  "app-id" : 525463029,
  "attribution-signature" : "MEYCIQDTuQ1Z4Tpy9D3aEKbxLl5J5iKiTumcqZikuY/AOD2U7QIhAJAaiAv89AoquHXJffcieEQXdWHpcV8ZgbKN0EwV9/sY",
  "redownload": true,
  "source-app-id": 1234567891,
  "fidelity-type": 1,
  "conversion-value": 20
}`

	v3_0__win = `{ 
  "version": "3.0", 
  "ad-network-id": "example123.skadnetwork", 
  "campaign-id": 42, 
  "transaction-id": "6aafb7a5-0170-41b5-bbe4-fe71dedf1e28", 
  "app-id": 525463029, 
  "attribution-signature": "MEYCIQD5eq3AUlamORiGovqFiHWI4RZT/PrM3VEiXUrsC+M51wIhAPMANZA9c07raZJ64gVaXhB9+9yZj/X6DcNxONdccQij", 
  "redownload": true, 
  "source-app-id": 1234567891, 
  "fidelity-type": 1, 
  "conversion-value": 20,
  "did-win": true
}`

	v3_0__lose = `{ 
  "version": "3.0",
  "ad-network-id": "example123.skadnetwork",
  "campaign-id": 42,
  "transaction-id": "f9ac267a-a889-44ce-b5f7-0166d11461f0",
  "app-id": 525463029,
  "attribution-signature": "MEUCIQDDetUtkyc/MiQvVJ5I6HIO1E7l598572Wljot2Onzd4wIgVJLzVcyAV+TXksGNoa0DTMXEPgNPeHCmD4fw1ABXX0g=",
  "redownload": true,
  "fidelity-type": 1,
  "did-win": false
}`
)

func ref[T any](t T) *T { return &t }

func TestSignAndVerify(t *testing.T) {
	s, err := skadnetwork.NewSigner(pem)
	assert.NilError(t, err)

	nonce := uuid.MustParse("68483ef6-0ada-40df-ab6b-3d19a66330fa")
	timestamp, _ := time.Parse(time.RFC3339, "2022-05-06T10:00:00Z")

	for _, c := range []struct {
		in *skadnetwork.Params
	}{
		{
			&skadnetwork.Params{
				AdNetworkID:      "example123.skadnetwork",
				CampaignID:       42,
				ItunesItemID:     525463029,
				Nonce:            nonce,
				SourceAppStoreID: 1234567891,
				Timestamp:        timestamp,
			},
		},
		{
			&skadnetwork.Params{
				AdNetworkID:      "example123.skadnetwork",
				CampaignID:       42,
				ItunesItemID:     525463029,
				Nonce:            nonce,
				SourceAppStoreID: 1234567891,
				Timestamp:        timestamp,
				FidelityType:     skadnetwork.SKRenderedAds,
			},
		},
		{
			&skadnetwork.Params{
				AdNetworkID:      "example123.skadnetwork",
				CampaignID:       42,
				ItunesItemID:     525463029,
				Nonce:            nonce,
				SourceAppStoreID: 1234567891,
				Timestamp:        timestamp,
				FidelityType:     skadnetwork.SKRenderedAds,
			},
		},
	} {
		sig, err := s.Sign(c.in)
		assert.NilError(t, err)

		got, err := s.Verify(c.in, sig)
		assert.NilError(t, err)
		assert.Equal(t, got, true)
	}
}

func TestMarshalJSON(t *testing.T) {
	for _, c := range []struct {
		in   string
		want *skadnetwork.Postback
	}{
		{
			v2_2,
			&skadnetwork.Postback{
				Version:              "2.2",
				AdNetworkID:          "com.example",
				CampaignID:           42,
				TransactionID:        "6aafb7a5-0170-41b5-bbe4-fe71dedf1e28",
				AppID:                525463029,
				AttributionSignature: "MEYCIQDTuQ1Z4Tpy9D3aEKbxLl5J5iKiTumcqZikuY/AOD2U7QIhAJAaiAv89AoquHXJffcieEQXdWHpcV8ZgbKN0EwV9/sY",
				Redownload:           ref(true),
				SourceAppID:          ref[int64](1234567891),
				FidelityType:         ref(skadnetwork.SKRenderedAds),
				ConversionValue:      ref[uint8](20),
			},
		},
		{
			v3_0__win,
			&skadnetwork.Postback{
				Version:              "3.0",
				AdNetworkID:          "example123.skadnetwork",
				CampaignID:           42,
				TransactionID:        "6aafb7a5-0170-41b5-bbe4-fe71dedf1e28",
				AppID:                525463029,
				AttributionSignature: "MEYCIQD5eq3AUlamORiGovqFiHWI4RZT/PrM3VEiXUrsC+M51wIhAPMANZA9c07raZJ64gVaXhB9+9yZj/X6DcNxONdccQij",
				Redownload:           ref(true),
				SourceAppID:          ref[int64](1234567891),
				FidelityType:         ref(skadnetwork.SKRenderedAds),
				ConversionValue:      ref[uint8](20),
				DidWin:               ref(true),
			},
		},
		{
			v3_0__lose,
			&skadnetwork.Postback{
				Version:              "3.0",
				AdNetworkID:          "example123.skadnetwork",
				CampaignID:           42,
				TransactionID:        "f9ac267a-a889-44ce-b5f7-0166d11461f0",
				AppID:                525463029,
				AttributionSignature: "MEUCIQDDetUtkyc/MiQvVJ5I6HIO1E7l598572Wljot2Onzd4wIgVJLzVcyAV+TXksGNoa0DTMXEPgNPeHCmD4fw1ABXX0g=",
				Redownload:           ref(true),
				FidelityType:         ref(skadnetwork.SKRenderedAds),
				DidWin:               ref(false),
			},
		},
	} {
		var got skadnetwork.Postback
		err := json.Unmarshal([]byte(c.in), &got)
		assert.NilError(t, err)

		_, err = json.Marshal(&got)
		assert.NilError(t, err)

		assert.Check(t, is.DeepEqual(&got, c.want))
	}
}

func TestVerifyPostback(t *testing.T) {
	for _, c := range []struct {
		in   string
		want bool
	}{
		{v2_2, true},
		{v3_0__win, true},
		{v3_0__lose, true},
	} {
		var p skadnetwork.Postback
		err := json.Unmarshal([]byte(c.in), &p)
		assert.NilError(t, err)

		got, err := skadnetwork.Verify(p)
		assert.NilError(t, err)

		assert.Equal(t, got, c.want)
	}
}
