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
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/service"
	"k8s.io/minikube/pkg/minikube/style"
)

func enableOrDisableNgrok(cfg *config.ClusterConfig, name, val string) error {
	enable, err := strconv.ParseBool(val)
	if err != nil {
		return errors.Wrapf(err, "parsing bool: %s", name)
	}
	if enable {
		return enableAddonNgrok(cfg)
	}
	// return disableAddonNgrok(cfg)
	return nil
}

func enableAddonNgrok(cfg *config.ClusterConfig) error {
	// out.Styled(style.Tip, "In enableAddonNgrok")
	exists, err := service.CheckSecretExists(cfg.Name, "ngrok-ingress-controller", "ngrok-ingress-controller-credentials")
	if err != nil {
		return errors.Wrap(err, "check secret exists")
	}
	if !exists {
		return fmt.Errorf("please run `minikube %s addons configure ngrok` to create your credentials before enabling", cfg.Name)
	}
	return nil
}

// func disableAddonNgrok(cfg *config.ClusterConfig) error {
// 	// TODO: Clean up existing secret
// 	return nil
// }

// ngrokValidation is a validator which checks if the ngrok credentials secret exists
// before enabling the ngrok addon. If not, it directs the user to run `minikube addons configure ngrok`
func ngrokValidation(cfg *config.ClusterConfig, _, _ string) error {
	exists, err := service.CheckSecretExists(cfg.Name, "ngrok-ingress-controller", "ngrok-ingress-controller-credentials")
	if err != nil {
		return errors.Wrap(err, "check secret exists")
	}
	if !exists {
		clusterName := cfg.Name
		tipProfileArg := ""
		if clusterName != constants.DefaultClusterName {
			tipProfileArg = fmt.Sprintf(" -p %s", clusterName)
		}
		out.Styled(style.Launch, fmt.Sprintf("Please run `minikube %s addons configure ngrok` to create your credentials before enabling the ngrok ingress addon", tipProfileArg))
		return ErrSkipThisAddon
	}
	return nil
}

// ngrokPostStart is a post-start callback which prints the ngrok URL
func ngrokPostStart(cfg *config.ClusterConfig, _, _ string) error {
	out.Styled(style.Notice, "The ngrok service is available at the following URL:")
	// AskForYesNoConfirmation("are you happy?", []string{"yes"}, []string{"no"})
	return nil
}

func AskForYesNoConfirmation(s string, posResponses, negResponses []string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		out.String("%s [y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		switch r := strings.ToLower(strings.TrimSpace(response)); {
		case containsString(posResponses, r):
			return true
		case containsString(negResponses, r):
			return false
		default:
			out.Err("Please type yes or no:")
		}
	}
}

// posString returns the first index of element in slice.
// If slice does not contain element, returns -1.
func posString(slice []string, element string) int {
	for index, elem := range slice {
		if elem == element {
			return index
		}
	}
	return -1
}

// containsString returns true if slice contains element
func containsString(slice []string, element string) bool {
	return posString(slice, element) != -1
}
