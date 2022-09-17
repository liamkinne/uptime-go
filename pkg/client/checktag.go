package uptime

import (
	"context"
	"encoding/json"
	"fmt"
)

// https://uptime.com/api/v1/docs/#/check-tags/get_service_tag_detail
type CheckTag struct {
	Pk       uint   `json:"pk,omitempty"`
	Url      string `json:"url,omitempty"`
	Name     string `json:"tag"`
	ColorHex string `json:"color_hex"`
}

// https://uptime.com/api/v1/docs/#/check-tags/get_servicetaglist
type CheckTagListResult struct {
	Count    uint       `json:"count"`
	Next     *uint      `json:"next"`
	Previous *uint      `json:"previous"`
	Results  []CheckTag `json:"results"`
}

// https://uptime.com/api/v1/docs/#/check-tags/post_servicetaglist
type CheckTagPostResult struct {
	Count    uint     `json:"count"`
	Next     *uint    `json:"next"`
	Previous *uint    `json:"previous"`
	Results  CheckTag `json:"results"`
}

func (c *UptimeClient) NewCheckTag(ctx context.Context, checkTag *CheckTag) (pk uint, err error) {
	tag := CheckTag{
		Name:     checkTag.Name,
		ColorHex: checkTag.ColorHex,
	}

	res, err := c.post(ctx, "check-tags/", tag)
	if err != nil {
		return 0, err
	}

	if res == nil {
		return 0, fmt.Errorf("no response body")
	}

	result := CheckTagPostResult{}

	err = json.Unmarshal(res, &result)
	if err != nil {
		return 0, err
	}

	return result.Results.Pk, err
}

func (c *UptimeClient) GetCheckTag(ctx context.Context, pk uint) (*CheckTag, error) {
	var checkTag *CheckTag

	err := c.get(ctx, fmt.Sprintf("check-tags/%d", pk), &checkTag, nil)
	if err != nil {
		return nil, err
	}

	return checkTag, nil
}

func (c *UptimeClient) GetCheckTags(ctx context.Context) ([]CheckTag, error) {
	var checkTags *CheckTagListResult

	err := c.get(ctx, "check-tags/", &checkTags, nil)
	if err != nil {
		return nil, err
	}

	return checkTags.Results, nil
}

func (c *UptimeClient) UpdateCheckTag(ctx context.Context, checkTag *CheckTag) error {
	tag := CheckTag{
		Name:     checkTag.Name,
		ColorHex: checkTag.ColorHex,
	}

	err := c.put(ctx, fmt.Sprintf("check-tags/%d", checkTag.Pk), tag)

	return err
}

func (c *UptimeClient) DeleteCheckTag(ctx context.Context, pk uint) error {
	err := c.delete(ctx, fmt.Sprintf("check-tags/%d", pk), nil)

	return err
}
