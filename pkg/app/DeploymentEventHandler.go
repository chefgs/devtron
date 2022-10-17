/*
 * Copyright (c) 2020 Devtron Labs
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
 *
 */

package app

import (
	"strings"
	"time"

	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	client "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	util "github.com/devtron-labs/devtron/util/event"
	"go.uber.org/zap"
)

type DeploymentEventHandler interface {
	WriteCDFailureEvent(pipelineId, appId, envId int)
	WriteCDSuccessEvent(pipelineId, appId, envId int)
}

type DeploymentEventHandlerImpl struct {
	logger            *zap.SugaredLogger
	appListingService AppListingService
	eventFactory      client.EventFactory
	eventClient       client.EventClient
}

func NewDeploymentEventHandlerImpl(logger *zap.SugaredLogger, appListingService AppListingService, eventClient client.EventClient, eventFactory client.EventFactory) *DeploymentEventHandlerImpl {
	deploymentEventHandlerImpl := &DeploymentEventHandlerImpl{
		logger:            logger,
		appListingService: appListingService,
		eventClient:       eventClient,
		eventFactory:      eventFactory,
	}
	return deploymentEventHandlerImpl
}

func (impl *DeploymentEventHandlerImpl) WriteCDFailureEvent(pipelineId, appId, envId int) {
	event := impl.eventFactory.Build(util.Fail, &pipelineId, appId, &envId, util.CD)
	impl.logger.Debugw("event WriteCDFailureEvent", "event", event)
	event = impl.eventFactory.BuildExtraCDData(event, nil, 0, bean.CD_WORKFLOW_TYPE_DEPLOY)
	_, evtErr := impl.eventClient.WriteNotificationEvent(event)
	if evtErr != nil {
		impl.logger.Errorw("error in writing event", "err", evtErr)
	}
}

func (impl *DeploymentEventHandlerImpl) WriteCDSuccessEvent(pipelineId, appId, envId int) {
	event := impl.eventFactory.Build(util.Success, &pipelineId, appId, &envId, util.CD)
	impl.logger.Debugw("event WriteCDSuccessEvent", "event", event)
	event = impl.eventFactory.BuildExtraCDData(event, nil, 0, bean.CD_WORKFLOW_TYPE_DEPLOY)
	_, evtErr := impl.eventClient.WriteNotificationEvent(event)
	if evtErr != nil {
		impl.logger.Errorw("error in writing event", "err", evtErr)
	}
}

func (impl *DeploymentEventHandlerImpl) BuildPayload(appName string, deploymentFailureTime time.Time) *client.Payload {
	applicationName := appName[:strings.LastIndex(appName, "-")]
	evnName := appName[strings.LastIndex(appName, "-")+1:]
	payload := &client.Payload{}
	payload.AppName = applicationName
	payload.EnvName = evnName
	//payload["deploymentFailureTime"] = deploymentFailureTime.Format(bean.LayoutRFC3339)
	return payload
}

func (impl *DeploymentEventHandlerImpl) isDeploymentFailed(ds repository.DeploymentStatus) bool {
	return ds.Status == application.Degraded && time.Since(ds.UpdatedOn) > 5*time.Minute
}