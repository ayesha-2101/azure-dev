package repository

import (
	"context"
	"fmt"

	"github.com/azure/azure-dev/cli/azd/internal/appdetect"
	"github.com/azure/azure-dev/cli/azd/internal/tracing"
	"github.com/azure/azure-dev/cli/azd/internal/tracing/fields"
	"github.com/azure/azure-dev/cli/azd/pkg/input"
	"github.com/azure/azure-dev/cli/azd/pkg/output"
	"github.com/fatih/color"
	"go.opentelemetry.io/otel/attribute"
)

type ServiceTargetKind string

const (
	NonSpecifiedTarget       ServiceTargetKind = ""
	AppServiceTarget         ServiceTargetKind = "appservice"
	ContainerAppTarget       ServiceTargetKind = "containerapp"
	AzureFunctionTarget      ServiceTargetKind = "function"
	StaticWebAppTarget       ServiceTargetKind = "staticwebapp"
	SpringAppTarget          ServiceTargetKind = "springapp"
	AksTarget                ServiceTargetKind = "aks"
	DotNetContainerAppTarget ServiceTargetKind = "appservice-dotnet"
	AiEndpointTarget         ServiceTargetKind = "ai.endpoint"
)

// detectConfirmAppHost handles prompting for confirming the detected project with an app host.
type detectConfirmAppHost struct {
	// The app host we found
	AppHost appdetect.Project

	// the root directory of the project
	root string

	// internal state and components
	console       input.Console
	targetService ServiceTargetKind
}

// Init initializes state from initial detection output
func (d *detectConfirmAppHost) Init(appHost appdetect.Project, root string) {
	fmt.Println("Select the target service for running aspire app:")
	fmt.Println("1. App Services")
	fmt.Println("2. Container Apps")

	var selection int
	_, err := fmt.Scan(&selection)
	if err != nil {
		fmt.Println("Error reading selection:", err)
		return
	}

	switch selection {
	case 1:
		d.AppHost = appHost
		d.targetService = AppServiceTarget
	case 2:
		d.AppHost = appHost
		d.targetService = ContainerAppTarget
	default:
		fmt.Println("Invalid selection. Defaulting to App Services.")
		d.AppHost = appHost
		d.targetService = AppServiceTarget
	}

	d.captureUsage(
		fields.AppInitDetectedServices)
}

func (d *detectConfirmAppHost) captureUsage(
	services attribute.Key) {

	tracing.SetUsageAttributes(
		services.StringSlice([]string{string(d.AppHost.Language)}),
	)
}

// Confirm prompts the user to confirm the detected services and databases,
// providing modifications to the detected services and databases.
func (d *detectConfirmAppHost) Confirm(ctx context.Context) (ServiceTargetKind, error) {
	for {
		if err := d.render(ctx); err != nil {
			return "", err
		}

		continueOption, err := d.console.Select(ctx, input.ConsoleOptions{
			Message: "Select an option",
			Options: []string{
				"Confirm and continue initializing my app",
				"Cancel and exit",
			},
		})
		if err != nil {
			return "", err
		}

		switch continueOption {
		case 0:
			d.captureUsage(
				fields.AppInitConfirmedServices)
			return d.targetService, nil
		case 1:
			return "", fmt.Errorf("cancelled due to user input")
		}
	}
}

func (d *detectConfirmAppHost) render(ctx context.Context) error {
	d.console.Message(ctx, "\n"+output.WithBold("Detected services:")+"\n")

	d.console.Message(ctx, "  "+color.BlueString(projectDisplayName(d.AppHost)))
	d.console.Message(ctx, "  "+"Detected in: "+output.WithHighLightFormat(relSafe(d.root, d.AppHost.Path)))
	d.console.Message(ctx, "")
	if d.targetService == "appservice" {
		d.console.Message(
			ctx,
			"azd will generate the files necessary to host your app on Azure using "+color.MagentaString(
				"Azure App Service",
			)+".\n",
		)
	} else if d.targetService == "containerapp" {
		d.console.Message(
			ctx,
			"azd will generate the files necessary to host your app on Azure using "+color.MagentaString(
				"Azure Container Apps",
			)+".\n",
		)
	}
	return nil
}
