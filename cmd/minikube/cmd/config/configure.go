/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package config

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/addons"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/service"
	"k8s.io/minikube/pkg/minikube/style"
)

var posResponses = []string{"yes", "y"}
var negResponses = []string{"no", "n"}

var addonsConfigureCmd = &cobra.Command{
	Use:   "configure ADDON_NAME",
	Short: "Configures the addon w/ADDON_NAME within minikube (example: minikube addons configure registry-creds). For a list of available addons use: minikube addons list",
	Long:  "Configures the addon w/ADDON_NAME within minikube (example: minikube addons configure registry-creds). For a list of available addons use: minikube addons list",
	Run: func(_ *cobra.Command, args []string) {
		if len(args) != 1 {
			exit.Message(reason.Usage, "usage: minikube addons configure ADDON_NAME")
		}

		profile := ClusterFlagValue()

		addon := args[0]
		// allows for additional prompting of information when enabling addons
		switch addon {
		case "registry-creds":

			// Default values
			awsAccessID := "changeme"
			awsAccessKey := "changeme"
			awsSessionToken := ""
			awsRegion := "changeme"
			awsAccount := "changeme"
			awsRole := "changeme"
			gcrApplicationDefaultCredentials := "changeme"
			dockerServer := "changeme"
			dockerUser := "changeme"
			dockerPass := "changeme"
			gcrURL := "https://gcr.io"
			acrURL := "changeme"
			acrClientID := "changeme"
			acrPassword := "changeme"

			enableAWSECR := AskForYesNoConfirmation("\nDo you want to enable AWS Elastic Container Registry?", posResponses, negResponses)
			if enableAWSECR {
				awsAccessID = AskForStaticValue("-- Enter AWS Access Key ID: ")
				awsAccessKey = AskForStaticValue("-- Enter AWS Secret Access Key: ")
				awsSessionToken = AskForStaticValueOptional("-- (Optional) Enter AWS Session Token: ")
				awsRegion = AskForStaticValue("-- Enter AWS Region: ")
				awsAccount = AskForStaticValue("-- Enter 12 digit AWS Account ID (Comma separated list): ")
				awsRole = AskForStaticValueOptional("-- (Optional) Enter ARN of AWS role to assume: ")
			}

			enableGCR := AskForYesNoConfirmation("\nDo you want to enable Google Container Registry?", posResponses, negResponses)
			if enableGCR {
				gcrPath := AskForStaticValue("-- Enter path to credentials (e.g. /home/user/.config/gcloud/application_default_credentials.json):")
				gcrchangeURL := AskForYesNoConfirmation("-- Do you want to change the GCR URL (Default https://gcr.io)?", posResponses, negResponses)

				if gcrchangeURL {
					gcrURL = AskForStaticValue("-- Enter GCR URL (e.g. https://asia.gcr.io):")
				}

				// Read file from disk
				dat, err := os.ReadFile(gcrPath)

				if err != nil {
					out.FailureT("Error reading {{.path}}: {{.error}}", out.V{"path": gcrPath, "error": err})
				} else {
					gcrApplicationDefaultCredentials = string(dat)
				}
			}

			enableDR := AskForYesNoConfirmation("\nDo you want to enable Docker Registry?", posResponses, negResponses)
			if enableDR {
				dockerServer = AskForStaticValue("-- Enter docker registry server url: ")
				dockerUser = AskForStaticValue("-- Enter docker registry username: ")
				dockerPass = AskForPasswordValue("-- Enter docker registry password: ")
			}

			enableACR := AskForYesNoConfirmation("\nDo you want to enable Azure Container Registry?", posResponses, negResponses)
			if enableACR {
				acrURL = AskForStaticValue("-- Enter Azure Container Registry (ACR) URL: ")
				acrClientID = AskForStaticValue("-- Enter client ID (service principal ID) to access ACR: ")
				acrPassword = AskForPasswordValue("-- Enter service principal password to access Azure Container Registry: ")
			}

			namespace := "kube-system"

			// Create ECR Secret
			err := service.CreateSecret(
				profile,
				namespace,
				"registry-creds-ecr",
				map[string]string{
					"AWS_ACCESS_KEY_ID":     awsAccessID,
					"AWS_SECRET_ACCESS_KEY": awsAccessKey,
					"AWS_SESSION_TOKEN":     awsSessionToken,
					"aws-account":           awsAccount,
					"aws-region":            awsRegion,
					"aws-assume-role":       awsRole,
				},
				map[string]string{
					"app":                           "registry-creds",
					"cloud":                         "ecr",
					"kubernetes.io/minikube-addons": "registry-creds",
				})
			if err != nil {
				out.FailureT("ERROR creating `registry-creds-ecr` secret: {{.error}}", out.V{"error": err})
			}

			// Create GCR Secret
			err = service.CreateSecret(
				profile,
				namespace,
				"registry-creds-gcr",
				map[string]string{
					"application_default_credentials.json": gcrApplicationDefaultCredentials,
					"gcrurl":                               gcrURL,
				},
				map[string]string{
					"app":                           "registry-creds",
					"cloud":                         "gcr",
					"kubernetes.io/minikube-addons": "registry-creds",
				})

			if err != nil {
				out.FailureT("ERROR creating `registry-creds-gcr` secret: {{.error}}", out.V{"error": err})
			}

			// Create Docker Secret
			err = service.CreateSecret(
				profile,
				namespace,
				"registry-creds-dpr",
				map[string]string{
					"DOCKER_PRIVATE_REGISTRY_SERVER":   dockerServer,
					"DOCKER_PRIVATE_REGISTRY_USER":     dockerUser,
					"DOCKER_PRIVATE_REGISTRY_PASSWORD": dockerPass,
				},
				map[string]string{
					"app":                           "registry-creds",
					"cloud":                         "dpr",
					"kubernetes.io/minikube-addons": "registry-creds",
				})

			if err != nil {
				out.WarningT("ERROR creating `registry-creds-dpr` secret")
			}

			// Create Azure Container Registry Secret
			err = service.CreateSecret(
				profile,
				namespace,
				"registry-creds-acr",
				map[string]string{
					"ACR_URL":       acrURL,
					"ACR_CLIENT_ID": acrClientID,
					"ACR_PASSWORD":  acrPassword,
				},
				map[string]string{
					"app":                           "registry-creds",
					"cloud":                         "acr",
					"kubernetes.io/minikube-addons": "registry-creds",
				})

			if err != nil {
				out.WarningT("ERROR creating `registry-creds-acr` secret")
			}

		case "metallb":
			_, cfg := mustload.Partial(profile)

			validator := func(s string) bool {
				return net.ParseIP(s) != nil
			}

			cfg.KubernetesConfig.LoadBalancerStartIP = AskForStaticValidatedValue("-- Enter Load Balancer Start IP: ", validator)

			cfg.KubernetesConfig.LoadBalancerEndIP = AskForStaticValidatedValue("-- Enter Load Balancer End IP: ", validator)

			if err := config.SaveProfile(profile, cfg); err != nil {
				out.ErrT(style.Fatal, "Failed to save config {{.profile}}", out.V{"profile": profile})
			}

			// Re-enable metallb addon in order to generate template manifest files with Load Balancer Start/End IP
			if err := addons.EnableOrDisableAddon(cfg, "metallb", "true"); err != nil {
				out.ErrT(style.Fatal, "Failed to configure metallb IP {{.profile}}", out.V{"profile": profile})
			}
		case "ingress":
			_, cfg := mustload.Partial(profile)

			validator := func(s string) bool {
				format := regexp.MustCompile("^.+/.+$")
				return format.MatchString(s)
			}

			customCert := AskForStaticValidatedValue("-- Enter custom cert (format is \"namespace/secret\"): ", validator)
			if cfg.KubernetesConfig.CustomIngressCert != "" {
				overwrite := AskForYesNoConfirmation("A custom cert for ingress has already been set. Do you want overwrite it?", posResponses, negResponses)
				if !overwrite {
					break
				}
			}

			cfg.KubernetesConfig.CustomIngressCert = customCert

			if err := config.SaveProfile(profile, cfg); err != nil {
				out.ErrT(style.Fatal, "Failed to save config {{.profile}}", out.V{"profile": profile})
			}
		case "registry-aliases":
			_, cfg := mustload.Partial(profile)
			validator := func(s string) bool {
				format := regexp.MustCompile(`^([a-zA-Z0-9-_]+\.[a-zA-Z0-9-_]+)+(\ [a-zA-Z0-9-_]+\.[a-zA-Z0-9-_]+)*$`)
				return format.MatchString(s)
			}
			registryAliases := AskForStaticValidatedValue("-- Enter registry aliases separated by space: ", validator)
			cfg.KubernetesConfig.RegistryAliases = registryAliases

			if err := config.SaveProfile(profile, cfg); err != nil {
				out.ErrT(style.Fatal, "Failed to save config {{.profile}}", out.V{"profile": profile})
			}
			addon := assets.Addons["registry-aliases"]
			if addon.IsEnabled(cfg) {
				// Re-enable registry-aliases addon in order to generate template manifest files with custom hosts
				if err := addons.EnableOrDisableAddon(cfg, "registry-aliases", "true"); err != nil {
					out.ErrT(style.Fatal, "Failed to configure registry-aliases {{.profile}}", out.V{"profile": profile})
				}
			}
		case "auto-pause":
			_, cfg := mustload.Partial(profile)
			intervalInput := AskForStaticValue("-- Enter interval time of auto-pause-interval (ex. 1m0s): ")
			intervalTime, err := time.ParseDuration(intervalInput)
			if err != nil {
				out.ErrT(style.Fatal, "Interval is an invalid duration: {{.error}}", out.V{"error": err})
			}
			if intervalTime != intervalTime.Abs() || intervalTime.String() == "0s" {
				out.ErrT(style.Fatal, "Interval must be greater than 0s")
			}
			cfg.AutoPauseInterval = intervalTime
			if err := config.SaveProfile(profile, cfg); err != nil {
				out.ErrT(style.Fatal, "Failed to save config {{.profile}}", out.V{"profile": profile})
			}
			addon := assets.Addons["auto-pause"]
			if addon.IsEnabled(cfg) {
				// Re-enable auto-pause addon in order to update interval time
				if err := addons.EnableOrDisableAddon(cfg, "auto-pause", "true"); err != nil {
					out.ErrT(style.Fatal, "Failed to configure auto-pause {{.profile}}", out.V{"profile": profile})
				}
			}
		case "ngrok":
			namespace := "ngrok-ingress-controller"
			// check if namespace exists first
			// if not, create namespace
			exists, err := service.CheckNamespace(profile, namespace)
			if err != nil {
				out.ErrT(style.Fatal, "Error checking for existing `ngrok-ingress-controller` namespace: {{.error}}", out.V{"error": err})
			}
			if !exists {
				err := service.CreateNamespace(profile, namespace)
				if err != nil {
					out.ErrT(style.Fatal, "Error creating `ngrok-ingress-controller` namespace: {{.error}}", out.V{"error": err})
				}
			}

			exists, err = service.CheckSecretExists(profile, namespace, "ngrok-ingress-controller-credentials")
			if err != nil {
				out.ErrT(style.Fatal, "Error checking for existing `ngrok-ingress-controller-credentials secret`: {{.error}}", out.V{"error": err})
			}

			replaceOrSet := "set"
			if exists {
				replaceOrSet = "replace"
			}
			msg := fmt.Sprintf("Would you like to %s ngrok credentials?", replaceOrSet)
			setCredentials := AskForYesNoConfirmation(msg, posResponses, negResponses)
			if setCredentials {
				ngrokAuthToken := AskForStaticValue("-- Enter ngrok authtoken: ")
				ngrokAPIKey := AskForStaticValue("-- Enter ngrok apikey: ")

				// Create ngrok Secret
				err := service.CreateSecret(
					profile,
					namespace,
					"ngrok-ingress-controller-credentials",
					map[string]string{
						"AUTHTOKEN": ngrokAuthToken,
						"API_KEY":   ngrokAPIKey,
					},
					map[string]string{
						"app":                           "ngrok",
						"cloud":                         "ngrok",
						"kubernetes.io/minikube-addons": "ngrok",
					})
				if err != nil {
					out.FailureT("ERROR creating `ngrok-ingress-controller-credentials` secret: {{.error}}", out.V{"error": err})
				}
				// TODO: Restart the container if credentials were updated
			}

			configureIngress := AskForYesNoConfirmation("Would you like to configure ingress for existing services in your minikube cluster?", posResponses, negResponses)
			if !configureIngress {
				break
			}

			domainChoice := AskForSingleOrMultipleDomainChoice()
			singleDomain := false
			if domainChoice == "single" {
				singleDomain = true
			}

			var domain string
			if singleDomain {
				domain = AskForStaticValue("What domain would you like to use?")
			}

			configurePolicyModule := AskForYesNoConfirmation("Would you like to create a policy module that can be used to secure ingress to services with authentication?", posResponses, negResponses)
			if configurePolicyModule {
				LoadAndCreatePolicyModules()
			}

			configureIngressToServices := AskForYesNoConfirmation("Would you like to create ingress to existing services?", posResponses, negResponses)
			if configureIngressToServices {
				AddIngressToServices(profile, singleDomain, domain)
			}

			out.SuccessT("Congrats, you have configured ingress with ngrok in your minikube cluster.")

			// TODO: Restart the container if credentials were updated
		default:
			out.FailureT("{{.name}} has no available configuration options", out.V{"name": addon})
			return
		}

		out.SuccessT("{{.name}} was successfully configured", out.V{"name": addon})
	},
}

func AskForSingleOrMultipleDomainChoice() string {
	choice := AskForChoice("Would you like to use a single domain for all services or a unique domain for each? Free accounts can only use a single domain.", []string{"single", "multiple"})
	return choice
}

func LoadAndCreatePolicyModules() {
	for {
		policyModulePath := AskForStaticValueOptional("Give path to policy module file (type 'none' if done):")
		if policyModulePath == "none" {
			break
		}
		// Load the policy module file and create the policy module
		_, err := ReadFileContents(policyModulePath)
		if err != nil {
			out.Err("Error reading module set file: %v", err)
			continue
		}
		// TODO: Take the output of the file and create a ngrok module set with a policy module
		// there will likely be issues here because we would need to load the ngrok crd definition
		// into the go client i think. maybe thats just apimachinery and the go client just takes strings
		// though. we will see.

	}
}

func AddIngressToServices(profile string, singleDomain bool, domain string) {
	// List all services in the cluster
	services, err := service.ListAllServices(profile)
	if err != nil {
		out.Err("Error listing services: %v", err)
		return
	}
	serviceString := ""
	for _, service := range services.Items {
		for _, port := range service.Spec.Ports {
			serviceString += fmt.Sprintf("%s:%s:%d\n", service.Namespace, service.Name, port.Port)
		}
	}
	out.Boxed("Services:\n" + serviceString)

	for {
		serviceChoice := AskForStaticValue("What service would you like to add ingress for? Use the format namespace:service:port (type 'none' to exit):")
		if serviceChoice == "none" {
			break
		}
		// Check if the service exists in the list of services
		found := false
		for _, service := range services.Items {
			for _, port := range service.Spec.Ports {
				if fmt.Sprintf("%s:%s:%d", service.Namespace, service.Name, port.Port) == serviceChoice {
					found = true
					break
				}
			}
		}
		if !found {
			out.Err("Service not found")
			continue
		}

		// create ingress object variable to be modified by following functions and applied at the end

		// addModuleSet := AskForYesNoConfirmation("Add a module set for authentication?", posResponses, negResponses)
		// if addModuleSet {
		// 	moduleSetName := AskForStaticValue("What module set would you like to add?")
		// 	// Check if the module set exists and add it to the ingress
		// 	// This is a placeholder; you'll need to implement the logic to add the module set to the ingress
		// }

		if !singleDomain {
			domain = AskForStaticValue("What domain would you like to use?")
		}

		serviceParts := strings.Split(serviceChoice, ":")
		namespace := serviceParts[0]
		serviceName := serviceParts[1]
		port, err := strconv.Atoi(string(serviceParts[2]))
		if err != nil {
			out.Err("Error converting port to int: %v", err)
			return
		}
		// Create the ingress object
		err = service.CreateIngress(profile, namespace, "ngrok-ingress-"+serviceName, domain, serviceName, int32(port), "ngrok")
		if err != nil {
			out.Err("Error creating ingress: %v", err)
		}

	}
}

func ReadFileContents(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file '%s': %w", filePath, err)
	}
	return string(data), nil
}

func init() {
	AddonsCmd.AddCommand(addonsConfigureCmd)
}
