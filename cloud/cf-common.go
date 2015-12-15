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
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/trustedanalytics/application-broker/misc"
	"github.com/trustedanalytics/application-broker/misc/http-utils"
	"net/http"
)

func (c *CfAPI) deleteEntity(url string, entityName string) error {
	log.Infof("Deleting %s: %v", entityName, url)

	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.Do(request)
	if err != nil {
		msg := fmt.Sprintf("Could not delete %s: [%v]", entityName, err)
		log.Error(msg)
		return misc.InternalServerError{Context: msg}
	}
	log.Debugf("Delete %s response code: %d", entityName, resp.StatusCode)

	if resp.StatusCode == http.StatusNotFound {
		log.Infof("%v already does not exist: %v", entityName, url)
	} else if !httputils.IsSuccessStatus(resp.StatusCode) {
		msg := fmt.Sprintf("Delete %s failed. Response from CC: (%d) [%v]",
			entityName, resp.StatusCode, misc.ReaderToString(resp.Body))
		log.Error(msg)
		return misc.InternalServerError{Context: msg}
	}

	return nil
}

func (c *CfAPI) getEntity(url string, entityName string) (*http.Response, error) {
	log.Infof("Getting %s: %v", entityName, url)

	response, err := c.Get(url)
	if err != nil {
		msg := fmt.Sprintf("Could not get %s: [%v]", entityName, err)
		log.Error(msg)
		return nil, misc.InternalServerError{Context: msg}
	}

	if response.StatusCode == http.StatusNotFound {
		return nil, misc.EntityNotFoundError{}
	}

	if response.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("Get %s failed. Response from CC: (%d) [%v]",
			entityName, response.StatusCode, misc.ReaderToString(response.Body))
		log.Error(msg)
		return nil, misc.InternalServerError{Context: msg}
	}

	return response, nil
}