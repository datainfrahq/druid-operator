package lookup

import (
	"encoding/json"
	"fmt"
	internalhttp "github.com/datainfrahq/druid-operator/pkg/http"
	"net/http"
	"time"
)

type Cluster struct {
	baseUrl    string
	httpClient internalhttp.DruidHTTP
}

func New(baseUrl string, httpClient internalhttp.DruidHTTP) (*Cluster, error) {
	cluster := Cluster{
		baseUrl:    baseUrl,
		httpClient: httpClient,
	}

	if err := cluster.initialize(); err != nil {
		return nil, err
	}

	return &cluster, nil
}

func (c *Cluster) Reconcile(desiredLookups map[LookupKey]string) error {
	actualLookups, err := c.listAll()
	if err != nil {
		return err
	}

	lookupsToDelete := make(map[LookupKey]struct{})
	for key := range actualLookups {
		if _, found := desiredLookups[key]; !found {
			lookupsToDelete[key] = struct{}{}
		}
	}

	for key := range lookupsToDelete {
		if err := c.delete(key.Tier, key.Id); err != nil {
			return err
		}
	}

	for key, spec := range desiredLookups {
		if err := c.upsert(key.Tier, key.Id, spec); err != nil {
			return err
		}
	}

	return nil
}

func (c *Cluster) initialize() error {
	url := fmt.Sprintf("%s/druid/coordinator/v1/lookups/config", c.baseUrl)
	_, err := c.httpClient.Do(http.MethodPost, url, []byte("{}"))
	return err
}

func (c *Cluster) listAll() (map[LookupKey]struct{}, error) {
	url := fmt.Sprintf("%s/druid/coordinator/v1/lookups/config/all", c.baseUrl)

	resp, err := c.httpClient.Do(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var response map[string]map[string]struct {
		version                string
		lookupExtractorFactory interface{}
	}
	if err := json.Unmarshal([]byte(resp.ResponseBody), &response); err != nil {
		return nil, err
	}

	lookups := make(map[LookupKey]struct{})
	for tier, lookupsInTier := range response {
		for id := range lookupsInTier {
			key := LookupKey{
				Tier: tier,
				Id:   id,
			}
			lookups[key] = struct{}{}
		}
	}

	return lookups, nil
}

func (c *Cluster) upsert(tier string, id string, spec string) error {
	if tier == "" {
		tier = "__default"
	}

	url := fmt.Sprintf("%s/druid/coordinator/v1/lookups/config/%s/%s", c.baseUrl, tier, id)

	body := fmt.Sprintf(
		`{ "version": "%s", "lookupExtractorFactory": %s }`,
		time.Now().Format(time.RFC3339),
		spec,
	)

	_, err := c.httpClient.Do(http.MethodPost, url, []byte(body))
	return err
}

func (c *Cluster) delete(tier string, id string) error {
	if tier == "" {
		tier = "__default"
	}

	url := fmt.Sprintf("%s/druid/coordinator/v1/lookups/config/%s/%s", c.baseUrl, tier, id)

	_, err := c.httpClient.Do(http.MethodDelete, url, nil)
	return err
}
