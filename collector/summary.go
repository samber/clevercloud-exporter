package collector

import (
	"context"
	"strconv"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"go.clever-cloud.dev/client"
	"go.clever-cloud.dev/sdk"
	"go.clever-cloud.dev/sdk/models"
	"go.clever-cloud.dev/sdk/v2/organisation"
)

type DeploymentCollector struct {
	client *client.Client
	mutex  sync.Mutex

	appStatus   *prometheus.Desc
	addonStatus *prometheus.Desc

	instanceStatus           *prometheus.Desc
	instanceMemoryTotal      *prometheus.Desc
	instanceCPUTotal         *prometheus.Desc
	instanceHourlyPriceTotal *prometheus.Desc
}

func NewDeploymentCollector(namespace string, client *client.Client) *DeploymentCollector {
	logrus.Info("CleverCloud collector enabled")

	return &DeploymentCollector{
		client: client,
		mutex:  sync.Mutex{},

		// applications
		appStatus: prometheus.NewDesc(
			namespace+"_application_status",
			"Applications status",
			[]string{"org_id", "org_name", "app_id", "app_name", "region", "archived", "type"}, nil,
		),

		// addons
		addonStatus: prometheus.NewDesc(
			namespace+"_addon_status",
			"Addons status",
			[]string{"org_id", "org_name", "addon_id", "addon_name", "region"}, nil,
		),

		// instances
		instanceStatus: prometheus.NewDesc(
			namespace+"_instance_status",
			"Instances status",
			[]string{"org_id", "org_name", "app_id", "app_name", "region", "type", "state"}, nil,
		),
		instanceMemoryTotal: prometheus.NewDesc(
			namespace+"_instance_memory_total",
			"Instances memory total",
			[]string{"org_id", "org_name", "app_id", "app_name", "region", "type"}, nil,
		),
		instanceCPUTotal: prometheus.NewDesc(
			namespace+"_instance_cpu_total",
			"Instances CPU total",
			[]string{"org_id", "org_name", "app_id", "app_name", "region", "type"}, nil,
		),
		instanceHourlyPriceTotal: prometheus.NewDesc(
			namespace+"_instance_hourly_price_total",
			"Instances price total",
			[]string{"org_id", "org_name", "app_id", "app_name", "region", "type"}, nil,
		),
	}
}

func (c *DeploymentCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.appStatus
	ch <- c.addonStatus
	ch <- c.instanceStatus
	ch <- c.instanceMemoryTotal
	ch <- c.instanceCPUTotal
	ch <- c.instanceHourlyPriceTotal
}

func (c *DeploymentCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	orgs, err := parse(organisation.List(context.Background(), c.client))
	if err != nil {
		logrus.Error(err)
		return
	}

	for _, org := range *orgs {
		// buffer all instances
		allInstances, errInstance := parse(ListInstances(context.Background(), c.client, org.ID))
		if errInstance != nil {
			logrus.Error(err)
		}

		// apps
		apps, err := parse(organisation.ListApplications(context.Background(), c.client, org.ID))
		if err != nil {
			logrus.Error(err)
		} else {
			for _, app := range *apps {
				ch <- prometheus.MustNewConstMetric(
					c.appStatus,
					prometheus.GaugeValue,
					float64(lo.Ternary(app.State == "SHOULD_BE_UP", 1, 0)),
					org.ID,
					org.Name,
					app.ID,
					app.Name,
					app.Zone,
					strconv.FormatBool(app.Archived),
					app.Instance.Type,
				)

				// instances
				if errInstance == nil {
					if instances, ok := (*allInstances)[app.ID]; ok {
						perState := lo.CountValuesBy(instances, func(item models.SuperNovaInstanceView) string {
							return item.State
						})

						for state, count := range perState {
							ch <- prometheus.MustNewConstMetric(
								c.instanceStatus,
								prometheus.GaugeValue,
								float64(count),
								org.ID,
								org.Name,
								app.ID,
								app.Name,
								app.Zone,
								app.Instance.Type,
								state,
							)
						}

						ch <- prometheus.MustNewConstMetric(
							c.instanceMemoryTotal,
							prometheus.GaugeValue,
							lo.SumBy(instances, func(item models.SuperNovaInstanceView) float64 {
								return float64(item.Flavor.Mem)
							}),
							org.ID,
							org.Name,
							app.ID,
							app.Name,
							app.Zone,
							app.Instance.Type,
						)
						ch <- prometheus.MustNewConstMetric(
							c.instanceCPUTotal,
							prometheus.GaugeValue,
							lo.SumBy(instances, func(item models.SuperNovaInstanceView) float64 {
								return float64(item.Flavor.Cpus)
							}),
							org.ID,
							org.Name,
							app.ID,
							app.Name,
							app.Zone,
							app.Instance.Type,
						)
						ch <- prometheus.MustNewConstMetric(
							c.instanceHourlyPriceTotal,
							prometheus.GaugeValue,
							lo.SumBy(instances, func(item models.SuperNovaInstanceView) float64 {
								return float64(item.Flavor.Price)
							}),
							org.ID,
							org.Name,
							app.ID,
							app.Name,
							app.Zone,
							app.Instance.Type,
						)
					}
				}
			}
		}

		// addons
		addons, err := parse(ListAddons(context.Background(), c.client, org.ID))
		if err != nil {
			logrus.Error(err)
		} else {
			for _, addon := range *addons {
				ch <- prometheus.MustNewConstMetric(
					c.addonStatus,
					prometheus.GaugeValue,
					1,
					org.ID,
					org.Name,
					addon.ID,
					addon.Name,
					addon.Region,
				)
			}
		}
	}
}

func parse[T any](response client.Response[T]) (*T, error) {
	if response.HasError() {
		return nil, response.Error()
	}

	return response.Payload(), nil
}

func ListAddons(ctx context.Context, cc *client.Client, orgID string, parameters ...sdk.Parameter) client.Response[[]models.AddonView] {
	url := sdk.Tpl("/v2/organisations/{id}/addons", map[string]string{"id": orgID})
	res := client.Get[[]models.AddonView](ctx, cc, url)
	return res
}

func ListInstances(ctx context.Context, cc *client.Client, orgID string, parameters ...sdk.Parameter) client.Response[map[string][]models.SuperNovaInstanceView] {
	url := sdk.Tpl("/v2/organisations/{id}/instances", map[string]string{"id": orgID})
	res := client.Get[map[string][]models.SuperNovaInstanceView](ctx, cc, url)
	return res
}
