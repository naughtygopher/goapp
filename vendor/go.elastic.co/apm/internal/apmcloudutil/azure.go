// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package apmcloudutil

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"

	"go.elastic.co/apm/model"
)

const (
	azureMetadataURL = "http://169.254.169.254/metadata/instance/compute?api-version=2019-08-15"
)

// See: https://docs.microsoft.com/en-us/azure/virtual-machines/windows/instance-metadata-service
func getAzureCloudMetadata(ctx context.Context, client *http.Client, out *model.Cloud) error {
	// First check for Azure App Service environment variables, which can
	// be done without performing any network requests.
	if getAzureAppServiceCloudMetadata(ctx, out) {
		return nil
	}

	req, err := http.NewRequest("GET", azureMetadataURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Metadata", "true")

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	var azureMetadata struct {
		Location          string `json:"location"`
		Name              string `json:"name"`
		ResourceGroupName string `json:"resourceGroupName"`
		SubscriptionID    string `json:"subscriptionId"`
		VMID              string `json:"vmId"`
		VMSize            string `json:"vmSize"`
		Zone              string `json:"zone"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&azureMetadata); err != nil {
		return err
	}

	out.Region = azureMetadata.Location
	out.AvailabilityZone = azureMetadata.Zone
	if azureMetadata.VMID != "" || azureMetadata.Name != "" {
		out.Instance = &model.CloudInstance{ID: azureMetadata.VMID, Name: azureMetadata.Name}
	}
	if azureMetadata.VMSize != "" {
		out.Machine = &model.CloudMachine{Type: azureMetadata.VMSize}
	}
	if azureMetadata.ResourceGroupName != "" {
		out.Project = &model.CloudProject{Name: azureMetadata.ResourceGroupName}
	}
	if azureMetadata.SubscriptionID != "" {
		out.Account = &model.CloudAccount{ID: azureMetadata.SubscriptionID}
	}
	return nil
}

func getAzureAppServiceCloudMetadata(ctx context.Context, out *model.Cloud) bool {
	// WEBSITE_OWNER_NAME has the form:
	//    {subscription id}+{app service plan resource group}-{region}webspace{.*}
	websiteOwnerName := os.Getenv("WEBSITE_OWNER_NAME")
	if websiteOwnerName == "" {
		return false
	}
	websiteInstanceID := os.Getenv("WEBSITE_INSTANCE_ID")
	if websiteInstanceID == "" {
		return false
	}
	websiteSiteName := os.Getenv("WEBSITE_SITE_NAME")
	if websiteSiteName == "" {
		return false
	}
	websiteResourceGroup := os.Getenv("WEBSITE_RESOURCE_GROUP")
	if websiteResourceGroup == "" {
		return false
	}

	plus := strings.IndexRune(websiteOwnerName, '+')
	if plus == -1 {
		return false
	}
	out.Account = &model.CloudAccount{ID: websiteOwnerName[:plus]}
	websiteOwnerName = websiteOwnerName[plus+1:]

	webspace := strings.LastIndex(websiteOwnerName, "webspace")
	if webspace == -1 {
		return false
	}
	websiteOwnerName = websiteOwnerName[:webspace]

	hyphen := strings.LastIndex(websiteOwnerName, "-")
	if hyphen == -1 {
		return false
	}
	out.Region = websiteOwnerName[hyphen+1:]
	out.Instance = &model.CloudInstance{ID: websiteInstanceID, Name: websiteSiteName}
	out.Project = &model.CloudProject{Name: websiteResourceGroup}
	return true
}
