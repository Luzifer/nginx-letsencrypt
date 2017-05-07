package main

import (
	"bufio"
	"os"
	"regexp"
	"sort"
	"strings"
)

var serverLine = regexp.MustCompile(`server_name ([^;]+);`)

func collectServerNames() ([]string, error) {
	serverNames := []string{}

	for _, filename := range cfg.NginxConfigs {
		if filename == "" {
			continue
		}

		f, err := os.Open(filename)
		if err != nil {
			return serverNames, err
		}

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if serverLine.MatchString(scanner.Text()) {
				serverNames = append(serverNames, strings.Split(serverLine.FindStringSubmatch(scanner.Text())[1], " ")...)
			}
		}
	}

	return serverNames, nil
}

func collectServerNameGroups(in []string, err error) (map[string][]string, error) {
	if err != nil {
		return nil, err
	}

	res := map[string][]string{}
	for _, fqdn := range in {
		sec := secondLevelDomain(fqdn)
		if _, ok := res[sec]; !ok {
			res[sec] = []string{}
		}

		r := res[sec]
		r = append(r, fqdn)
		r = domainSort(r)
		res[sec] = r
	}

	return res, nil
}

func domainSort(in []string) []string {
	sort.Slice(in, func(i, j int) bool {
		return reversedFQDN(in[i]) < reversedFQDN(in[j])
	})
	return in
}

func secondLevelDomain(fqdn string) string {
	rev := reversedFQDN(fqdn)
	rev2nd := strings.Join(strings.Split(rev, ".")[0:2], ".")
	second := reversedFQDN(rev2nd)

	return second
}

func reversedFQDN(fqdn string) string {
	originalParts := strings.Split(fqdn, ".")
	reversedParts := []string{}

	for i := len(originalParts) - 1; i >= 0; i-- {
		reversedParts = append(reversedParts, originalParts[i])
	}

	return strings.Join(reversedParts, ".")
}
