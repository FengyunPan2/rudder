package models

import(
	rls "k8s.io/helm/pkg/proto/hapi/services"
)

type GetReleaseResponse struct {
	Status      rls.GetReleaseStatusResponse
	Content     rls.GetReleaseContentResponse
	History     rls.GetHistoryResponse
}
