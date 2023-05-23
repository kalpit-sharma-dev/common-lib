package entitlement

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/coocood/freecache"
	apiModel "gitlab.kksharmadevdev.com/platform/platform-api-model/clients/model/Golang/resourceModel/entitlement"
)

//go:generate mockgen -package mock -destination=mock/mocks.go -source=service.go

type EntitlementService interface {
	GetPartnerFeatures(partnerID string) (features []apiModel.Feature, err error)
	IsPartnerAuthorized(partnerID, featureName string) bool
	GetPartnerFeatureNames(partnerID string) (featureNames map[string]bool, err error)
}

// Service represents an Entitlement Service type
type Service struct {
	cache           *freecache.Cache
	httpClient      *http.Client
	url             string
	cacheDataTTLSec int
}

// NewEntitlementService creates a new Entitlement Service
func NewEntitlementService(httpClient *http.Client, entitlementMsURL string, cacheDataTTLSec, cacheSize int) Service {
	return Service{
		cache:           freecache.NewCache(cacheSize),
		httpClient:      httpClient,
		url:             entitlementMsURL,
		cacheDataTTLSec: cacheDataTTLSec,
	}
}

// GetPartnerFeatures retrieve features for Partner from Entitlement MS or from local cache
func (es Service) GetPartnerFeatures(partnerID string) (features []apiModel.Feature, err error) {
	var featuresBin []byte
	partnerIDBin := []byte(partnerID)

	featuresBin, err = es.cache.Get(partnerIDBin)
	if err != nil {
		resp, entitlementErr := es.httpClient.Get(es.url + "/partners/" + partnerID + "/features")
		if entitlementErr != nil {
			return features, entitlementErr
		}
		defer resp.Body.Close()

		featuresBin, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return features, err
		}

		err = es.cache.Set(partnerIDBin, featuresBin, es.cacheDataTTLSec)

		if err != nil {
			return features, err
		}
	}

	err = json.Unmarshal(featuresBin, &features)

	return features, err
}

func (es Service) GetPartnerFeatureNames(partnerID string) (featureNames map[string]bool, err error) {

	key := []byte(partnerID)
	val, err := es.cache.Get(key)

	if err != nil {
		// Get feature list from REST call
		features, err := es.getFeaturesRest(partnerID)
		if err != nil || len(features) < 1 {
			return featureNames, err
		}

		// Populate feature name map
		featureNames = make(map[string]bool)
		for _, feature := range features {
			featureNames[feature.Name] = true
		}

		// Store feature name map in cache
		featuresBin, err := json.Marshal(featureNames)
		if err == nil {
			es.cache.Set(key, featuresBin, es.cacheDataTTLSec)
		}
		return featureNames, err
	}

	err = json.Unmarshal(val, &featureNames)

	return featureNames, err
}

// IsPartnerAuthorized checks if the Partner has enabled feature in the Entitlement Service
func (es Service) IsPartnerAuthorized(partnerID, featureName string) bool {
	features, err := es.GetPartnerFeatureNames(partnerID)
	if err != nil {
		return false
	}

	if ok := features[featureName]; ok {
		return true
	}

	return false
}

func (es Service) getFeaturesRest(partnerID string) (features []apiModel.Feature, err error) {
	resp, entitlementErr := es.httpClient.Get(es.url + "/partners/" + partnerID + "/features")
	if entitlementErr != nil {
		return nil, entitlementErr
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return features, err
	}

	json.Unmarshal(data, &features)

	return features, err
}
