// Copyright 2023 Terramate GmbH
// SPDX-License-Identifier: MPL-2.0

package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/rs/zerolog/log"
	"github.com/terramate-io/terramate/cloud"
	"github.com/terramate-io/terramate/cloud/deployment"
	"github.com/terramate-io/terramate/errors"
	prj "github.com/terramate-io/terramate/project"
)

func (c *cli) createCloudDeployment(runStacks []ExecContext) {
	logger := log.With().
		Logger()

	if !c.cloudEnabled() {
		return
	}

	logger = logger.With().
		Str("organization", c.cloud.run.orgUUID).
		Logger()

	ctx, cancel := context.WithTimeout(context.Background(), defaultCloudTimeout)
	defer cancel()

	var (
		err                 error
		deploymentCommitSHA string
		deploymentURL       string
		ghRepo              string
	)

	if c.prj.isRepo {
		r, err := repository.Parse(c.prj.prettyRepo())
		if err != nil {
			logger.Debug().
				Msg("repository cannot be normalized: skipping pull request retrievals for commit")
		} else {
			ghRepo = r.Owner + "/" + r.Name
		}

		deploymentCommitSHA = c.prj.headCommit()
	}

	ghRunID := os.Getenv("GITHUB_RUN_ID")
	ghAttempt := os.Getenv("GITHUB_RUN_ATTEMPT")
	if ghRunID != "" && ghAttempt != "" && ghRepo != "" {
		deploymentURL = fmt.Sprintf(
			"https://github.com/%s/actions/runs/%s/attempts/%s",
			ghRepo,
			ghRunID,
			ghAttempt,
		)

		logger.Debug().
			Str("deployment_url", deploymentURL).
			Msg("detected deployment url")
	}

	payload := cloud.DeploymentStacksPayloadRequest{
		ReviewRequest: c.cloud.run.reviewRequest,
		Workdir:       prj.PrjAbsPath(c.rootdir(), c.wd()),
		Metadata:      c.cloud.run.metadata,
	}

	for _, run := range runStacks {
		tags := run.Stack.Tags
		if tags == nil {
			tags = []string{}
		}
		payload.Stacks = append(payload.Stacks, cloud.DeploymentStackRequest{
			Stack: cloud.Stack{
				MetaID:          strings.ToLower(run.Stack.ID),
				MetaName:        run.Stack.Name,
				MetaDescription: run.Stack.Description,
				MetaTags:        tags,
				Repository:      c.prj.prettyRepo(),
				Path:            run.Stack.Dir.String(),
			},
			CommitSHA:         deploymentCommitSHA,
			DeploymentCommand: strings.Join(run.Cmd, " "),
			DeploymentURL:     deploymentURL,
		})
	}
	res, err := c.cloud.client.CreateDeploymentStacks(ctx, c.cloud.run.orgUUID, c.cloud.run.runUUID, payload)
	if err != nil {
		logger.Error().
			Err(err).
			Msg("failed to create cloud deployment")

		c.disableCloudFeatures(cloudError())
		return
	}

	if len(res) != len(runStacks) {
		logger.Error().
			Msgf("the backend respond with an invalid number of stacks in the deployment: %d instead of %d",
				len(res), len(runStacks))

		c.disableCloudFeatures(cloudError())
		return
	}

	for _, r := range res {
		logger.Debug().Msgf("deployment created: %+v\n", r)
		if r.StackMetaID == "" {
			logger.Error().
				Msg("backend returned empty meta_id")

			c.disableCloudFeatures(cloudError())
			return
		}
		c.cloud.run.meta2id[r.StackMetaID] = r.StackID
	}
}

func (c *cli) cloudSyncDeployment(runContext ExecContext, err error) {
	var status deployment.Status
	switch {
	case err == nil:
		status = deployment.OK
	case errors.IsKind(err, ErrRunCanceled):
		status = deployment.Canceled
	case errors.IsAnyKind(err, ErrRunFailed, ErrRunCommandNotFound):
		status = deployment.Failed
	default:
		panic(errors.E(errors.ErrInternal, "unexpected run status"))
	}

	c.doCloudSyncDeployment(runContext, status)
}

func (c *cli) doCloudSyncDeployment(runContext ExecContext, status deployment.Status) {
	st := runContext.Stack
	logger := log.With().
		Str("organization", c.cloud.run.orgUUID).
		Str("stack", st.RelPath()).
		Stringer("status", status).
		Logger()

	stackID, ok := c.cloud.run.meta2id[st.ID]
	if !ok {
		logger.Error().Msg("unable to update deployment status due to invalid API response")
		return
	}

	payload := cloud.UpdateDeploymentStacks{
		Stacks: []cloud.UpdateDeploymentStack{
			{
				StackID: stackID,
				Status:  status,
			},
		},
	}

	logger.Debug().Msg("updating deployment status")

	ctx, cancel := context.WithTimeout(context.Background(), defaultCloudTimeout)
	defer cancel()
	err := c.cloud.client.UpdateDeploymentStacks(ctx, c.cloud.run.orgUUID, c.cloud.run.runUUID, payload)
	if err != nil {
		logger.Err(err).Str("stack_id", st.ID).Msg("failed to update deployment status for each")
	} else {
		logger.Debug().Msg("deployment status synced successfully")
	}
}
