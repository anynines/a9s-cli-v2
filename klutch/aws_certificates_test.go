package klutch

import (
	"errors"
	"net"
	"testing"
)

type fakeProvisioner struct {
	ns []string
}

func (f *fakeProvisioner) EnsureCertificate(domainName string, altNames []string, hostedZoneName string) (string, error) {
	return "", nil
}
func (f *fakeProvisioner) EnsureCNAMERecords(hostedZoneName string, records map[string]string) error {
	return nil
}
func (f *fakeProvisioner) GetHostedZoneNS(hostedZoneName string) ([]string, error) {
	if len(f.ns) == 0 {
		return nil, errors.New("no NS")
	}
	return f.ns, nil
}
func (f *fakeProvisioner) EnsureALBAliasRecord(hostedZoneName, recordName, albDNSName string) error {
	return nil
}
func (f *fakeProvisioner) EnsurePublicHostedZone(hostedZoneName, clusterName string) ([]string, error) {
	return f.GetHostedZoneNS(hostedZoneName)
}

func TestVerifyHostedZoneResolvableWaitsForDelegation(t *testing.T) {
	origLookup := lookupNSFunc
	origDelay := dnsDelegationPollDelay
	origRetries := dnsDelegationMaxRetries
	defer func() {
		lookupNSFunc = origLookup
		dnsDelegationPollDelay = origDelay
		dnsDelegationMaxRetries = origRetries
	}()

	callCount := 0
	lookupNSFunc = func(name string) ([]*net.NS, error) {
		callCount++
		if callCount < 2 {
			return nil, errors.New("not delegated yet")
		}
		return []*net.NS{{Host: "ns-1.awsdns.com."}}, nil
	}
	dnsDelegationPollDelay = 0
	dnsDelegationMaxRetries = 3

	prov := &fakeProvisioner{ns: []string{"ns-1.awsdns.com.", "ns-2.awsdns.net."}}

	verifyHostedZoneResolvable(prov, "example.com", "")

	if callCount < 2 {
		t.Fatalf("expected multiple lookup attempts before delegation")
	}
}
