package router

import (
	"github.com/easystack/rudder/src/api"
	"github.com/easystack/rudder/src/models"
	rls "k8s.io/helm/pkg/proto/hapi/services"
	"k8s.io/helm/pkg/repo"
	"k8s.io/helm/cmd/helm/search"
	"github.com/easystack/rudder/src/router/filter"

	restful "github.com/emicklei/go-restful"
)

func CreateHTTPRouter() *restful.Container {
	ac := api.NewAPIClient()
	wsContainer := restful.NewContainer()
	wsContainer.Filter(filter.LogRequestAndReponse)
	ws := new(restful.WebService)

	ws.Path("/api/v1").
		Consumes(restful.MIME_XML, restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	//repo
	ws.Route(ws.GET("/repos").To(ac.ListRepos).Writes([]*repo.RepoFile{}))
	//chart
	ws.Route(ws.GET("/charts").To(ac.ListCharts).Writes([]*search.Result{}))
	//release
	// POST /api/v1/releases
	ws.Route(ws.POST("/release").To(ac.InstallRelease).
		Doc("install release. defaults: namespace=default, version=latest.").
		Operation("installRelease").
		Reads(models.InstallReleaseRequest{}).
		Writes(rls.GetReleaseStatusResponse{}))

	// GET /api/v1/releases
	ws.Route(ws.GET("/releases").To(ac.ListReleases).
		Doc("list releases").
		Operation("listReleases").
		Writes([]*rls.ListReleasesResponse{}))

	// GET /api/v1/release/{release}
	ws.Route(ws.GET("/release/{release}").To(ac.GetRelease).
		Doc("get release").
		Operation("getRelease").
		Writes([]*models.GetReleaseResponse{}))

	// GET /api/v1/release/{release}/history
	ws.Route(ws.GET("/release/{release}").To(ac.GetReleaseHistory).
		Doc("get release history").
		Operation("GetReleaseHistory").
		Writes([]*rls.GetHistoryResponse{}))

	// GET /api/v1/release/{release}/status
	ws.Route(ws.GET("/release/{release}").To(ac.GetReleaseStatus).
		Doc("get release status").
		Operation("GetReleaseStatus").
		Writes([]*rls.GetReleaseStatusResponse{}))

	// GET /api/v1/release/{release}/content
	ws.Route(ws.GET("/release/{release}").To(ac.GetReleaseContent).
		Doc("get release content").
		Operation("GetReleaseContent").
		Writes([]*rls.GetReleaseContentResponse{}))

	// PATCH /api/v1/release/{release}
	ws.Route(ws.PATCH("/release/{release}").To(ac.UpdateRelease).
		Doc("update release").
		Operation("updateRelease").
		Writes([]*rls.GetReleaseStatusResponse{}))

	// DELETE /api/v1/releases/{release}
	ws.Route(ws.DELETE("/release/{release}").To(ac.DeleteRelease).
		Doc("uninstall release").
		Operation("uninstallRelease"))

	wsContainer.Add(ws)

	return wsContainer
}
