package client

import (
	"fmt"

	"github.com/easystack/rudder/src/models"
	helmRepos "github.com/easystack/rudder/src/service/handlers/repos"
	helmCharts "github.com/easystack/rudder/src/service/handlers/charts"
	helmReleases "github.com/easystack/rudder/src/service/handlers/releases"
	helm_env "k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/helm/helmpath"

	log "github.com/Sirupsen/logrus"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/repo"
	"k8s.io/helm/cmd/helm/search"
	"k8s.io/helm/pkg/helm/portforwarder"
	"k8s.io/helm/pkg/kube"
	rls "k8s.io/helm/pkg/proto/hapi/services"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/client-go/rest"
)

const (
	TillerNamespace   = "kube-system"
	TillerServiceName = "tiller-deploy"
	KubeContext       = ""
	TillerPort        = 44134
)

type HelmClient struct {
	namespace     string
	tillerHost    string
	client        helm.Interface
	settings      *helm_env.EnvSettings
}

// NewHelmClient returns the Helm implementation of data.Client
func NewHelmClient(namespace string, tillerHost string) *HelmClient {
	client := GetTillerCon(namespace, tillerHost)
	settings := GetSettings(namespace, tillerHost)
	return &HelmClient{
		namespace:   namespace,
		tillerHost:  tillerHost,
		client:      client,
		settings:    settings,
	}
}

// release
func (c *HelmClient) ListReleases(listRelease *models.ListRelease) (*rls.ListReleasesResponse, error) {
	return helmReleases.GetAllReleases(c.client, listRelease, true)
}

func (c *HelmClient) GetRelease(getRelease *models.GetReleaseRequest) (*models.GetReleaseResponse, error) {
	return helmReleases.GetRelease(c.client, getRelease)
}

func (c *HelmClient) GetReleaseHistory(getRelease *models.GetReleaseRequest) (*rls.GetHistoryResponse, error) {
	result, err := helmReleases.GetRelease(c.client, getRelease)
	return &result.History, err
}

func (c *HelmClient) GetReleaseStatus(getRelease *models.GetReleaseRequest) (*rls.GetReleaseStatusResponse, error) {
	result, err := helmReleases.GetRelease(c.client, getRelease)
	return &result.Status, err
}

func (c *HelmClient) GetReleaseContent(getRelease *models.GetReleaseRequest) (*rls.GetReleaseContentResponse, error) {
	result, err := helmReleases.GetRelease(c.client, getRelease)
	return &result.Content, err
}

func (c *HelmClient) InstallRelease(installRelease *models.InstallReleaseRequest) (*rls.GetReleaseStatusResponse, error) {
	return helmReleases.InstallRelease(c.client, c.settings, installRelease)
}

func (c *HelmClient) UpdateRelease(installRelease *models.UpdateRelease) (*rls.GetReleaseStatusResponse, error) {
	return helmReleases.UpdateRelease(c.client, c.settings, installRelease)
}

func (c *HelmClient) DeleteReleases(deleteRelease *models.DeleteRelease) (*rls.UninstallReleaseResponse, error) {
	return helmReleases.DeleteRelease(c.client, deleteRelease)
}

// chart
func (c *HelmClient) ListCharts(listChart *models.ListChart) ([]*search.Result, error) {
	return helmCharts.GetAllCharts(c.client, listChart)
}

// repo
func (c *HelmClient) ListRepos() (*repo.RepoFile, error) {
	return helmRepos.GetAllRepos(c.client)
}

func GetTillerCon(namespace string, tillerHost string) *helm.Client {
	tillerHost, err := setupConnection(namespace, tillerHost)
	if err != nil {
		log.Fatalf("can't connect tiller: %v", err)
	} else {
		log.Printf("Tiller SERVER: %q\n", tillerHost)
	}
	client := helm.NewClient(helm.Host(tillerHost))
	return client
}

func GetSettings(namespace string, tillerHost string) *helm_env.EnvSettings {
	settings := new(helm_env.EnvSettings)

	helmHomeTemp := helm_env.DefaultHelmHome()
	settings.Home = helmpath.Home(helmHomeTemp)
	settings.Debug = false
	settings.TillerNamespace = namespace
	settings.TillerHost = tillerHost
	settings.PlugDirs = settings.Home.Plugins()
	return settings
}

// getKubeClient is a convenience method for creating kubernetes config and client
// for a given kubeconfig context
func getKubeClient(context string) (*rest.Config, *internalclientset.Clientset, error) {
	config, err := kube.GetConfig(context).ClientConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("could not get kubernetes config for context '%s': %s", context, err)
	}
	client, err := internalclientset.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get kubernetes client: %s", err)
	}
	return config, client, nil
}

func setupConnection(namespace string, tillerHost string) (string, error) {
	if namespace =="" {
		namespace = TillerNamespace
	}
	if tillerHost == "" {
		config, client, err := getKubeClient(KubeContext)
		if err != nil {
			return "", err
		}

		tunnel, err := portforwarder.New(namespace, client, config)
		if err != nil {
			return "", err
		}

		tillerHost = fmt.Sprintf("localhost:%d", tunnel.Local)
		//debug("Created tunnel using local port: '%d'\n", tunnel.Local)
	}

	// Set up the gRPC config.
	//debug("SERVER: %q\n", settings.TillerHost)

	// Plugin support.
	return tillerHost, nil
}
