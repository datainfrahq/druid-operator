package lookup

import (
	"encoding/json"
	"fmt"
	internalhttp "github.com/datainfrahq/druid-operator/pkg/http"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"reflect"
	"time"
)

type DruidClient struct {
	baseUrl    string
	httpClient internalhttp.DruidHTTP
}

func NewCluster(baseUrl string, httpClient internalhttp.DruidHTTP) (*DruidClient, error) {
	cluster := DruidClient{
		baseUrl:    baseUrl,
		httpClient: httpClient,
	}

	if err := cluster.initialize(); err != nil {
		return nil, err
	}

	return &cluster, nil
}

func (c *DruidClient) Reconcile(desiredLookups map[LookupKey]Spec, reports map[types.NamespacedName]Report) error {
	actualLookups, err := c.listAll()
	if err != nil {
		return err
	}

	lookupsToUpdate := make(map[LookupKey]Spec)
	lookupsToDelete := make(map[LookupKey]interface{})
	for key, desiredSpec := range desiredLookups {
		if actualSpec, found := actualLookups[key]; found {
			delete(actualLookups, key) // do not consider this key when we loop through actual lookups
			if reflect.DeepEqual(desiredSpec.spec, actualSpec) {
				continue
			} // desired and actual lookup match, no need to update
		}
		lookupsToUpdate[key] = desiredSpec
	}
	for key, actualSpec := range actualLookups {
		lookupsToDelete[key] = actualSpec
	}

	for key, spec := range lookupsToUpdate {
		report := Report(NewSuccessReport(v1.LocalObjectReference{Name: "placeholder"}, key.Tier, key.Id, spec.spec))
		if err := c.upsert(key.Tier, key.Id, spec.spec); err != nil {
			report = NewErrorReport(err)
		}
		if _, ok := reports[spec.name]; !ok {
			reports[spec.name] = report
		}
	}

	for key := range lookupsToDelete {
		if err := c.delete(key.Tier, key.Id); err != nil {
			return err
		}
	}

	return nil
}

func (c *DruidClient) GetStatus() (map[LookupKey]Status, error) {
	url := fmt.Sprintf("%s/druid/coordinator/v1/lookups/status?detailed=true", c.baseUrl)

	resp, err := c.httpClient.Do(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var response map[string]map[string]Status
	if err := json.Unmarshal([]byte(resp.ResponseBody), &response); err != nil {
		return nil, err
	}

	statues := make(map[LookupKey]Status)
	for tier, lookupsInTier := range response {
		for id, status := range lookupsInTier {
			key := LookupKey{
				Tier: tier,
				Id:   id,
			}
			statues[key] = status
		}
	}

	return statues, nil
}

func (c *DruidClient) initialize() error {
	url := fmt.Sprintf("%s/druid/coordinator/v1/lookups/config", c.baseUrl)
	_, err := c.httpClient.Do(http.MethodPost, url, []byte("{}"))
	return err
}

func (c *DruidClient) listAll() (map[LookupKey]interface{}, error) {
	url := fmt.Sprintf("%s/druid/coordinator/v1/lookups/config/all", c.baseUrl)

	resp, err := c.httpClient.Do(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var response map[string]map[string]struct {
		Version                string      `json:"version"`
		LookupExtractorFactory interface{} `json:"lookupExtractorFactory"`
	}
	if err := json.Unmarshal([]byte(resp.ResponseBody), &response); err != nil {
		return nil, err
	}

	lookups := make(map[LookupKey]interface{})
	for tier, lookupsInTier := range response {
		for id, spec := range lookupsInTier {
			key := LookupKey{
				Tier: tier,
				Id:   id,
			}
			lookups[key] = spec.LookupExtractorFactory
		}
	}

	return lookups, nil
}

func (c *DruidClient) upsert(tier string, id string, spec interface{}) error {
	if tier == "" {
		tier = "__default"
	}

	url := fmt.Sprintf("%s/druid/coordinator/v1/lookups/config/%s/%s", c.baseUrl, tier, id)

	spec, err := json.Marshal(spec)
	if err != nil {
		return err
	}

	body := fmt.Sprintf(
		`{ "version": "%s", "lookupExtractorFactory": %s }`,
		time.Now().Format(time.RFC3339),
		spec,
	)

	_, err = c.httpClient.Do(http.MethodPost, url, []byte(body))
	return err
}

func (c *DruidClient) delete(tier string, id string) error {
	if tier == "" {
		tier = "__default"
	}

	url := fmt.Sprintf("%s/druid/coordinator/v1/lookups/config/%s/%s", c.baseUrl, tier, id)

	_, err := c.httpClient.Do(http.MethodDelete, url, nil)
	return err
}
