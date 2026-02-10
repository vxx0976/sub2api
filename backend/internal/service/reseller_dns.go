package service

import (
	"fmt"
	"net"
	"strings"
)

// verifyDNSTXT checks if the domain has a TXT record containing the expected token.
// The user should add a TXT record at _domain-verify.<domain> with value "domain-verify=<token>".
func verifyDNSTXT(domain, expectedToken string) (bool, error) {
	verifyDomain := fmt.Sprintf("_domain-verify.%s", domain)
	expectedValue := fmt.Sprintf("domain-verify=%s", expectedToken)

	records, err := net.LookupTXT(verifyDomain)
	if err != nil {
		// DNS lookup failure — domain may not have the TXT record yet
		return false, nil
	}

	for _, record := range records {
		if strings.TrimSpace(record) == expectedValue {
			return true, nil
		}
	}

	return false, nil
}
