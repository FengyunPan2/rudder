package charts

import (
	log "github.com/Sirupsen/logrus"

	"k8s.io/helm/pkg/helm/helmpath"
	"github.com/easystack/rudder/src/models"
	"k8s.io/helm/cmd/helm/search"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/repo"
	"os"
	"path/filepath"
)

const (
	localRepoIndexFilePath = "index.yaml"
	homeEnvVar             = "HELM_HOME"
	hostEnvVar             = "HELM_HOST"
	tillerNamespaceEnvVar  = "TILLER_NAMESPACE"
	// searchMaxScore suggests that any score higher than this is not considered a match.
	searchMaxScore         = 25
)

func GetAllCharts(helmclient helm.Interface, listChart *models.ListChart) ([]*search.Result, error) {
	log.Printf("Call GetAllCharts: %q", listChart)
	setDefaultValue(listChart)

	index, err := buildIndex(listChart)
	if err != nil {
		return nil, err
	}

	var res []*search.Result
	if len(listChart.Filter) == 0 {
		res = index.All()
	} else {
		res, err = index.Search(listChart.Filter, searchMaxScore, listChart.Regexp)
		if err != nil {
			return nil, nil
		}
	}

	search.SortScore(res)
	return res, nil

}

func setDefaultValue(listChart *models.ListChart) {
	if listChart.Version == "" {
		listChart.Version = ""
	}
}

func buildIndex(listChart *models.ListChart) (*search.Index, error) {
	home := helmpath.Home(homePath())
	// Load the repositories.yaml
	rf, err := repo.LoadRepositoriesFile(home.RepositoryFile())
	if err != nil {
		return nil, err
	}

	i := search.NewIndex()
	for _, re := range rf.Repositories {
		n := re.Name
		f := home.CacheIndex(n)
		ind, err := repo.LoadIndexFile(f)
		if err != nil {
			log.Printf("WARNING: Repo %q is corrupt or missing. Try 'helm repo update'.", n)
			continue
		}

		i.AddRepo(n, ind, listChart.Versions)
	}
	return i, nil
}

func homePath() string {
	s := os.ExpandEnv(defaultHelmHome())
	os.Setenv(homeEnvVar, s)
	return s
}

func defaultHelmHome() string {
	if home := os.Getenv(homeEnvVar); home != "" {
		return home
	}
	return filepath.Join(os.Getenv("HOME"), ".helm")
}