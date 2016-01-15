/**
 * Copyright (c) 2015 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/trustedanalytics/application-broker/misc"
	"github.com/trustedanalytics/application-broker/misc/http-utils"
	"github.com/trustedanalytics/application-broker/types"
	"net/http"
)

func (c *CfAPI) createServiceInstance(req *types.CfServiceInstanceCreateRequest) (*types.CfServiceInstanceCreateResponse, error) {
	address := c.BaseAddress + "/v2/service_instances?accepts_incomplete=false"
	log.Infof("Requesting service instance creation: %v", address)
	marshalled, err := json.Marshal(req)
	if err != nil {
		log.Errorf("Could not marshal CfServiceInstanceCreateRequest: [%+v]", req)
		return nil, misc.InternalServerError{Context: "Problem with marshalling request data"}
	}
	resp, err := c.Post(address, "application/json", bytes.NewReader(marshalled))
	if err != nil {
		log.Errorf("Could not create service instance: [%v]", err)
		return nil, misc.InternalServerError{Context: "Cloud Foundry API was not able to create service instance"}
	}
	if !(resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusAccepted) {
		// CF 2.07 returns HTTP 201, CF 2.22 returns HTTP 202
		log.Errorf("createServiceInstance failed. Response from CC: [%v]", misc.ReaderToString(resp.Body))
		return nil, misc.InternalServerError{Context: "Unacceptable response code from Cloud Foundry API after trying to create service instance"}
	}

	toReturn := new(types.CfServiceInstanceCreateResponse)
	json.NewDecoder(resp.Body).Decode(toReturn)
	log.Debugf("createServiceInstance status code: [%v]", resp.StatusCode)
	log.Debugf("createServiceInstance returned GUID: [%v]", toReturn.Meta.GUID)
	return toReturn, nil
}

func (c *CfAPI) createServiceBinding(req *types.CfServiceBindingCreateRequest) (*types.CfServiceBindingCreateResponse, error) {
	address := c.BaseAddress + "/v2/service_bindings"
	log.Infof("Requesting service binding creation: %v", address)
	marshalled, err := json.Marshal(req)
	if err != nil {
		log.Errorf("Could not marshal CfServiceInstanceCreateRequest: [%+v]", req)
		return nil, misc.InternalServerError{Context: "Problem with marshalling request data"}
	}
	resp, err := c.Post(address, "application/json", bytes.NewReader(marshalled))
	if err != nil {
		log.Errorf("Could not create service binding: [%v]", err)
		return nil, misc.InternalServerError{Context: "Cloud Foundry API was not able to create service binding"}
	}
	if resp.StatusCode != http.StatusCreated {
		log.Errorf("createServiceBinding failed. Response from CC: [%v]", misc.ReaderToString(resp.Body))
		return nil, misc.InternalServerError{Context: "Unacceptable response code from Cloud Foundry API after trying to create service binding"}
	}

	toReturn := new(types.CfServiceBindingCreateResponse)
	json.NewDecoder(resp.Body).Decode(toReturn)
	log.Debugf("createServiceBinding status code: [%v]", resp.StatusCode)
	log.Debugf("createServiceBinding returned GUID: [%v]", toReturn.Meta.GUID)
	return toReturn, nil
}

func (c *CfAPI) deleteServiceInstance(id string) error {
	address := fmt.Sprintf("%v/v2/service_instances/%v", c.BaseAddress, id)
	err := c.deleteEntity(address, "service instance")
	if err != nil {
		log.Errorf("Error deleting service instance %v", id)
		return err
	}
	return nil
}

func (c *CfAPI) getServiceOfName(name string) (*types.CfServiceResource, error) {
	address := fmt.Sprintf("%v/v2/services?q=label:%v", c.BaseAddress, name)
	resp, err := c.Get(address)

	if err != nil {
		log.Errorf("Could not get service of name provided: [%v]", err)
		return nil, misc.InternalServerError{Context: "Request CF for service with given name, failed"}
	}
	if resp.StatusCode != http.StatusOK {
		log.Errorf("Problem while getting service of specified name: [%v]", err)
		return nil, misc.InternalServerError{Context: "Wrong status code from CF API after trying to get specific service"}
	}

	resource := new(types.CfServicesResources)
	json.NewDecoder(resp.Body).Decode(resource)
	if resource.TotalResults > 0 {
		log.Debugf("Service with name [%v] found", name)
		return &resource.Resources[0], nil
	}
	return nil, nil
}

func (c *CfAPI) purgeService(serviceID string, serviceName string, servicePlansURL string) error {
	log.Infof("Purge service: [%v]", serviceID)
	resp, err := c.Get(c.BaseAddress + servicePlansURL)
	plans := new(types.CfServicePlansResources)
	json.NewDecoder(resp.Body).Decode(plans)

	for _, plan := range plans.Resources {
		address := fmt.Sprintf("%v/v2/service_plans/%v", c.BaseAddress, plan.Meta.GUID)
		if err := c.deleteEntity(address, "service plan"); err != nil {
			return err
		}
	}

	address := fmt.Sprintf("%v/v2/services/%v", c.BaseAddress, serviceID)
	err = c.deleteEntity(address, "service")
	if err != nil {
		msg := fmt.Sprintf("Could not delete service %s: [%v]", serviceName, err)
		log.Error(msg)
		return misc.InternalServerError{Context: msg}
	}
	log.Debugf("Delete service %s response code: %d", serviceName, resp.StatusCode)

	if resp.StatusCode == http.StatusNotFound {
		log.Infof("%v already does not exist", serviceName)
	} else if !httputils.IsSuccessStatus(resp.StatusCode) {
		msg := fmt.Sprintf("Delete %s failed. Response from CC: (%d) [%v]",
			serviceName, resp.StatusCode, misc.ReaderToString(resp.Body))
		log.Error(msg)
		return misc.InternalServerError{Context: msg}
	}
	return nil
}
