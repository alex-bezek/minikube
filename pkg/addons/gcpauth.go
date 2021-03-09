/*
Copyright 2021 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package addons

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strconv"

	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/detect"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
)

// enableOrDisableGCPAuth enables or disables the gcp-auth addon depending on the val parameter
func enableOrDisableGCPAuth(cfg *config.ClusterConfig, name string, val string) error {
	enable, err := strconv.ParseBool(val)
	if err != nil {
		return errors.Wrapf(err, "parsing bool: %s", name)
	}
	if enable {
		return enableAddonGCPAuth(cfg)
	}
	return disableAddonGCPAuth(cfg)
}

func enableAddonGCPAuth(cfg *config.ClusterConfig) error {
	if !Force && detect.IsOnGCE() {
		exit.Message(reason.InternalCredsNotFound, "It seems that you are running in GCE, which means authentication should work without the GCP Auth addon. If you would still like to authenticate using a credentials file, use the --force flag.")
	}

	// Grab command runner from running cluster
	cc := mustload.Running(cfg.Name)
	r := cc.CP.Runner

	// Grab credentials from where GCP would normally look
	ctx := context.Background()
	creds, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		exit.Message(reason.InternalCredsNotFound, "Could not find any GCP credentials. Either run `gcloud auth application-default login` or set the GOOGLE_APPLICATION_CREDENTIALS environment variable to the path of your credentials file.")
	}

	// Don't mount in empty credentials file
	if creds.JSON == nil {
		exit.Message(reason.InternalCredsNotFound, "Could not find any GCP credentials. Either run `gcloud auth application-default login` or set the GOOGLE_APPLICATION_CREDENTIALS environment variable to the path of your credentials file.")
	}

	f := assets.NewMemoryAssetTarget(creds.JSON, credentialsPath, "0444")

	err = r.Copy(f)
	if err != nil {
		return err
	}

	// First check if the project env var is explicitly set
	projectEnv := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectEnv != "" {
		f := assets.NewMemoryAssetTarget([]byte(projectEnv), projectPath, "0444")
		return r.Copy(f)
	}

	// We're currently assuming gcloud is installed and in the user's path
	project, err := exec.Command("gcloud", "config", "get-value", "project").Output()
	if err == nil && len(project) > 0 {
		f := assets.NewMemoryAssetTarget(bytes.TrimSpace(project), projectPath, "0444")
		return r.Copy(f)
	}

	out.WarningT("Could not determine a Google Cloud project, which might be ok.")
	out.Styled(style.Tip, `To set your Google Cloud project,  run: 

		gcloud config set project <project name>

or set the GOOGLE_CLOUD_PROJECT environment variable.`)

	// Copy an empty file in to avoid errors about missing files
	emptyFile := assets.NewMemoryAssetTarget([]byte{}, projectPath, "0444")
	return r.Copy(emptyFile)
}

func disableAddonGCPAuth(cfg *config.ClusterConfig) error {
	// Grab command runner from running cluster
	cc := mustload.Running(cfg.Name)
	r := cc.CP.Runner

	// Clean up the files generated when enabling the addon
	creds := assets.NewMemoryAssetTarget([]byte{}, credentialsPath, "0444")
	err := r.Remove(creds)
	if err != nil {
		return err
	}

	project := assets.NewMemoryAssetTarget([]byte{}, projectPath, "0444")
	err = r.Remove(project)
	if err != nil {
		return err
	}

	return nil
}

func verifyGCPAuthAddon(cc *config.ClusterConfig, name string, val string) error {
	enable, err := strconv.ParseBool(val)
	if err != nil {
		return errors.Wrapf(err, "parsing bool: %s", name)
	}
	err = verifyAddonStatusInternal(cc, name, val, "gcp-auth")

	if enable && err == nil {
		out.Styled(style.Notice, "Your GCP credentials will now be mounted into every pod created in the {{.name}} cluster.", out.V{"name": cc.Name})
		out.Styled(style.Notice, "If you don't want your credentials mounted into a specific pod, add a label with the `gcp-auth-skip-secret` key to your pod configuration.")
	}

	return err
}
