package api

import (
	"net/http"
	log "github.com/Sirupsen/logrus"

	"github.com/easystack/rudder/src/config"
	helmclient "github.com/easystack/rudder/src/service/client"
	"github.com/easystack/rudder/src/models"

	restful "github.com/emicklei/go-restful"
)

type apiClient struct {
	hClient *helmclient.HelmClient
}

func NewAPIClient() *apiClient {
	conf := config.GetConfig()
	hc := helmclient.NewHelmClient(conf.Namespace, conf.TillerHost)
	return &apiClient{
		hClient: hc,
	}
}

// release
func (ac *apiClient) ListReleases(req *restful.Request, resp *restful.Response) {
	log.Printf("Requst ListReleases: %q", req)
	listRelease := new(models.ListRelease)
	err := req.ReadEntity(listRelease)
	if err != nil {
		handleInternalError(resp, err)
		return
	}

	releases, err := ac.hClient.ListReleases(listRelease)
	if err != nil {
		handleInternalError(resp, err)
		return
	}
	resp.WriteHeaderAndEntity(http.StatusOK, releases)
}

func (ac *apiClient) GetRelease(req *restful.Request, resp *restful.Response) {
	log.Printf("Requst GetRelease: %q", req)
	getRelease := new(models.GetReleaseRequest)
	err := req.ReadEntity(getRelease)
	if err != nil {
		handleInternalError(resp, err)
		return
	}

	releases, err := ac.hClient.GetRelease(getRelease)
	if err != nil {
		handleInternalError(resp, err)
		return
	}
	resp.WriteHeaderAndEntity(http.StatusOK, releases)
}

func (ac *apiClient) GetReleaseHistory(req *restful.Request, resp *restful.Response) {
	log.Printf("Requst GetReleaseHistory: %q", req)
	getRelease := new(models.GetReleaseRequest)
	err := req.ReadEntity(getRelease)
	if err != nil {
		handleInternalError(resp, err)
		return
	}

	releases, err := ac.hClient.GetReleaseHistory(getRelease)
	if err != nil {
		handleInternalError(resp, err)
		return
	}
	resp.WriteHeaderAndEntity(http.StatusOK, releases)
}

func (ac *apiClient) GetReleaseStatus(req *restful.Request, resp *restful.Response) {
	log.Printf("Requst GetReleaseStatus: %q", req)
	getRelease := new(models.GetReleaseRequest)
	err := req.ReadEntity(getRelease)
	if err != nil {
		handleInternalError(resp, err)
		return
	}

	releases, err := ac.hClient.GetReleaseStatus(getRelease)
	if err != nil {
		handleInternalError(resp, err)
		return
	}
	resp.WriteHeaderAndEntity(http.StatusOK, releases)
}

func (ac *apiClient) GetReleaseContent(req *restful.Request, resp *restful.Response) {
	log.Printf("Requst GetReleaseContent: %q", req)
	getRelease := new(models.GetReleaseRequest)
	err := req.ReadEntity(getRelease)
	if err != nil {
		handleInternalError(resp, err)
		return
	}

	releases, err := ac.hClient.GetReleaseContent(getRelease)
	if err != nil {
		handleInternalError(resp, err)
		return
	}
	resp.WriteHeaderAndEntity(http.StatusOK, releases)
}

func (ac *apiClient) InstallRelease(req *restful.Request, resp *restful.Response) {
	log.Printf("Requst ListReleases: %q", req)
	installRelease := new(models.InstallReleaseRequest)
	err := req.ReadEntity(installRelease)
	if err != nil {
		handleInternalError(resp, err)
		return
	}

	releases, err := ac.hClient.InstallRelease(installRelease)
	if err != nil {
		handleInternalError(resp, err)
		return
	}
	resp.WriteHeaderAndEntity(http.StatusOK, releases)
}

func (ac *apiClient) UpdateRelease(req *restful.Request, resp *restful.Response) {
	log.Printf("Requst UpdateRelease: %q", req)
	updateRelease := new(models.UpdateRelease)
	err := req.ReadEntity(updateRelease)
	if err != nil {
		handleInternalError(resp, err)
		return
	}

	release, err := ac.hClient.UpdateRelease(updateRelease)
	if err != nil {
		handleInternalError(resp, err)
		return
	}
	resp.WriteHeaderAndEntity(http.StatusOK, release)
}

func (ac *apiClient) DeleteRelease(req *restful.Request, resp *restful.Response) {
	log.Printf("Requst DeleteReleases: %q", req)
	deleteRelease := new(models.DeleteRelease)
	err := req.ReadEntity(deleteRelease)
	if err != nil {
		handleInternalError(resp, err)
		return
	}

	_, err = ac.hClient.DeleteReleases(deleteRelease)
	if err != nil {
		handleInternalError(resp, err)
		return
	}
	//resp.WriteHeaderAndEntity(http.StatusOK, releases)
	resp.WriteHeader(http.StatusOK)
}

// chart
func (ac *apiClient) ListCharts(req *restful.Request, resp *restful.Response) {
	log.Printf("Requst ListCharts: %q", req)
	listChart := new(models.ListChart)
	err := req.ReadEntity(listChart)
	if err != nil {
		handleInternalError(resp, err)
		return
	}

	charts, err := ac.hClient.ListCharts(listChart)
	if err != nil {
		handleInternalError(resp, err)
		return
	}
	resp.WriteHeaderAndEntity(http.StatusOK, charts)
}

// repo
func (ac *apiClient) ListRepos(req *restful.Request, resp *restful.Response) {
	log.Printf("Requst ListRepos: %q", req)
	repos, err := ac.hClient.ListRepos()
	if err != nil {
		handleInternalError(resp, err)
		return
	}
	resp.WriteHeaderAndEntity(http.StatusOK, repos)
}