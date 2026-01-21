package dedibox

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/dedibox/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

const (
	defaultServerTimeout  = 60 * time.Minute
	defaultServiceTimeout = 120 * time.Minute
	retryInterval         = 30 * time.Second
)

// newAPI returns a new dedibox API
func newAPI(m any) *dedibox.API {
	return dedibox.NewAPI(meta.ExtractScwClient(m))
}

// newAPIWithZone returns a new API and the zone for a Create request
func newAPIWithZone(d *schema.ResourceData, m any) (*dedibox.API, scw.Zone, error) {
	api := dedibox.NewAPI(meta.ExtractScwClient(m))

	zone, err := meta.ExtractZone(d, m)
	if err != nil {
		return nil, "", err
	}

	return api, zone, nil
}

// NewAPIWithZoneAndID returns an API with zone and ID extracted from the state
func NewAPIWithZoneAndID(m any, id string) (*dedibox.API, zonal.ID, error) {
	api := dedibox.NewAPI(meta.ExtractScwClient(m))

	zone, ID, err := zonal.ParseID(id)
	if err != nil {
		return nil, zonal.ID{}, err
	}

	return api, zonal.NewID(zone, ID), nil
}
