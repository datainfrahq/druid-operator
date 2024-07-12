package lookup

import (
	"encoding/json"
	"fmt"
	"github.com/datainfrahq/druid-operator/controllers/lookup/report"
	internalhttp "github.com/datainfrahq/druid-operator/pkg/http"
	"net/http"
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

func (c *DruidClient) GetStatus(tier string, id string) (report.StatusReport, error) {
	url := fmt.Sprintf("%s/druid/coordinator/v1/lookups/status/%s/%s?detailed=true", c.baseUrl, tier, id)

	resp, err := c.httpClient.Do(http.MethodGet, url, nil)
	if err != nil {
		return report.StatusReport{}, err
	}

	var response report.StatusReport
	if err := json.Unmarshal([]byte(resp.ResponseBody), &response); err != nil {
		return report.StatusReport{}, err
	}

	return response, nil
}

func (c *DruidClient) initialize() error {
	url := fmt.Sprintf("%s/druid/coordinator/v1/lookups/config", c.baseUrl)
	_, err := c.httpClient.Do(http.MethodPost, url, []byte("{}"))
	return err
}

func (c *DruidClient) Upsert(tier string, id string, spec interface{}) error {
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

func (c *DruidClient) Delete(tier string, id string) error {
	if tier == "" {
		tier = "__default"
	}

	url := fmt.Sprintf("%s/druid/coordinator/v1/lookups/config/%s/%s", c.baseUrl, tier, id)

	_, err := c.httpClient.Do(http.MethodDelete, url, nil)
	return err
}
