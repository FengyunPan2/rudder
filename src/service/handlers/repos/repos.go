package repos

import (
	"os"
	"errors"
	"path/filepath"
	log "github.com/Sirupsen/logrus"

	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/repo"
)

const (
	localRepoIndexFilePath = "index.yaml"
	homeEnvVar             = "HELM_HOME"
	hostEnvVar             = "HELM_HOST"
	tillerNamespaceEnvVar  = "TILLER_NAMESPACE"
)

func GetAllRepos(helmclient helm.Interface) (*repo.RepoFile, error) {
	log.Printf("Call GetAllRepos")
	home := helmpath.Home(homePath())
	repos, err := repo.LoadRepositoriesFile(home.RepositoryFile())
	if err != nil {
		return nil, err
	}

	if len(repos.Repositories) == 0 {
		return nil, errors.New("no repositories to show")
	}
	return repos, nil
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
