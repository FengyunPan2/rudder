package releases

import (
	"errors"
	"strings"
	"fmt"
	"text/template"
	"google.golang.org/grpc"

	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/downloader"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/storage/driver"
	log "github.com/Sirupsen/logrus"
	rls "k8s.io/helm/pkg/proto/hapi/services"
	"k8s.io/helm/pkg/proto/hapi/chart"
	helm_env "k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/getter"

	"github.com/easystack/rudder/src/models"
	"k8s.io/helm/pkg/proto/hapi/release"
	"os"
	"path/filepath"
	"github.com/Masterminds/sprig"
	"bytes"
)

var settings helm_env.EnvSettings
var errReleaseRequired = errors.New("release name is required")

// SortBy defines sort operations.
type ListSort_SortBy int32
type ListSort_SortOrder int32

const (
	ListSort_UNKNOWN       ListSort_SortBy = 0
	ListSort_NAME          ListSort_SortBy = 1
	ListSort_LAST_RELEASED ListSort_SortBy = 2
	ListSort_ASC           ListSort_SortOrder = 0
	ListSort_DESC          ListSort_SortOrder = 1
)

// GetReleases returns all the existing releases in your cluster
func GetAllReleases(helmclient helm.Interface, listRelease *models.ListRelease, releasesEnabled bool) (*rls.ListReleasesResponse, error) {
	log.Printf("Call GetAllReleases: %q", listRelease)
	if !releasesEnabled {
		return nil, fmt.Errorf("Feature not enabled")
	}

	setListReleaseDefaultValue(listRelease)

	sortBy := ListSort_NAME
	if listRelease.ByDate {
		sortBy = ListSort_LAST_RELEASED
	}

	sortOrder := ListSort_ASC
	if listRelease.SortDesc {
		sortOrder = ListSort_DESC
	}

	stats := statusCodes(listRelease)

	res, err := helmclient.ListReleases(
		helm.ReleaseListLimit(listRelease.Limit),
		helm.ReleaseListOffset(listRelease.Offset),
		helm.ReleaseListFilter(listRelease.Filter),
		helm.ReleaseListSort(int32(sortBy)),
		helm.ReleaseListOrder(int32(sortOrder)),
		helm.ReleaseListStatuses(stats),
		helm.ReleaseListNamespace(listRelease.Namespace),
	)

	if err != nil {
		return nil, prettyError(err)
	}

	if len(res.Releases) == 0 {
		return nil, nil
	}

	return res, nil
}

func GetRelease(helmclient helm.Interface, getRelease *models.GetReleaseRequest) (*models.GetReleaseResponse, error) {
	log.Printf("Call GetRelease: %q", getRelease)
	if len(getRelease.Name) == 0 {
		return nil, errReleaseRequired
	}
	if getRelease.Max == 0 {
		getRelease.Max = 256
	}

	history, err := helmclient.ReleaseHistory(getRelease.Name, helm.WithMaxHistory(getRelease.Max))
	if err != nil {
		return nil, prettyError(err)
	}

	status, err := helmclient.ReleaseStatus(getRelease.Name, helm.StatusReleaseVersion(getRelease.Revision))
	if err != nil {
		return nil, prettyError(err)
	}

	content, err := helmclient.ReleaseContent(getRelease.Name, helm.ContentReleaseVersion(getRelease.Revision))
	if err != nil {
		return nil, prettyError(err)
	}

	release := new(models.GetReleaseResponse)
	release.History = *history
	release.Status = *status
	release.Content = *content

	return release, nil
}

func InstallRelease(helmclient helm.Interface, helm_settings *helm_env.EnvSettings, installRelease *models.InstallReleaseRequest) (*rls.GetReleaseStatusResponse, error) {
	log.Printf("Call InstallRelease: %q", installRelease)
	setInstallReleaseDefaultValue(installRelease)

	//set helm ENV settings
	settings = *helm_settings

	if installRelease.Chart == "" {
		return nil, errors.New("'install release' requires a chart name")
	}
	cp, err := locateChartPath(installRelease.Chart,
		installRelease.Version, installRelease.Verify, installRelease.Keyring)
	if err != nil {
		msg := fmt.Sprintf("'install release' failed to get local chart path: %v", err)
		return nil, errors.New(msg)
	}
	installRelease.Chart = cp
	log.Printf("CHART PATH: %s\n", installRelease.Chart)

	// If template is specified, try to run the template.
	if installRelease.NameTemplate != "" {
		installRelease.Name, err = generateName(installRelease.NameTemplate)
		if err != nil {
			return nil, err
		}
		// Print the final name so the user knows what the final name of the release is.
		log.Printf("FINAL NAME: %s\n", installRelease.Name)
	}

	// Check chart requirements to make sure all dependencies are present in /charts
	chartRequested, err := chartutil.Load(installRelease.Chart)
	if err != nil {
		return nil, prettyError(err)
	}

	if req, err := chartutil.LoadRequirements(chartRequested); err == nil {
		// If checkDependencies returns an error, we have unfullfilled dependencies.
		// As of Helm 2.4.0, this is treated as a stopping condition:
		// https://github.com/kubernetes/helm/issues/2209
		if err := checkDependencies(chartRequested, req); err != nil {
			return nil, prettyError(err)
		}
	}

	rawVals := new([]byte)
	rel, err := helmclient.InstallReleaseFromChart(
		chartRequested,
		installRelease.Namespace,
		helm.ValueOverrides(*rawVals),
		helm.ReleaseName(installRelease.Name),
		helm.InstallDryRun(installRelease.DryRun),
		helm.InstallReuseName(installRelease.Replace),
		helm.InstallDisableHooks(installRelease.DisableHooks),
		helm.InstallTimeout(installRelease.Timeout),
		helm.InstallWait(installRelease.Wait))
	if err != nil {
		return nil, prettyError(err)
	}

	log.Printf("InstallRelease success! Happy Helming!\n")
	// Print the status like status command does
	release := rel.GetRelease()
	if release == nil {
		return nil, nil
	}
	status, err := helmclient.ReleaseStatus(release.Name)
	if err != nil {
		return nil, prettyError(err)
	}
	return status, nil
}

func UpdateRelease(helmclient helm.Interface, helm_settings *helm_env.EnvSettings, updateRelease *models.UpdateRelease) (*rls.GetReleaseStatusResponse, error) {
	log.Printf("Call UpdateRelease: %q", updateRelease)
	if updateRelease.Rollback {
		return rollbackRelease(helmclient, helm_settings, updateRelease)
	} else {
		return upgradeRelease(helmclient, helm_settings, updateRelease)
	}
}

func rollbackRelease(helmclient helm.Interface, helm_settings *helm_env.EnvSettings, updateRelease *models.UpdateRelease) (*rls.GetReleaseStatusResponse, error) {
	log.Printf("Call rollbackRelease: %q", updateRelease)
	setRollbackReleaseDefaultValue(updateRelease)
	if len(updateRelease.Release) == 0 {
		return nil, errors.New("'rollback release' requires a release name")
	}
	updateRelease.Revision = int32(updateRelease.Revision)
	rel, err := helmclient.RollbackRelease(
		updateRelease.Release,
		helm.RollbackDryRun(updateRelease.DryRun),
		helm.RollbackRecreate(updateRelease.Recreate),
		helm.RollbackDisableHooks(updateRelease.DisableHooks),
		helm.RollbackVersion(updateRelease.Revision),
		helm.RollbackTimeout(updateRelease.Timeout),
		helm.RollbackWait(updateRelease.Wait))
	if err != nil {
		return nil, prettyError(err)
	}

	log.Printf("Release %q has been rollbacked. Happy Helming!\n", updateRelease.Release)
	// Print the status like status command does
	release := rel.GetRelease()
	if release == nil {
		return nil, nil
	}
	status, err := helmclient.ReleaseStatus(release.Name)
	if err != nil {
		return nil, prettyError(err)
	}
	return status, nil
}

func upgradeRelease(helmclient helm.Interface, helm_settings *helm_env.EnvSettings, updateRelease *models.UpdateRelease) (*rls.GetReleaseStatusResponse, error) {
	log.Printf("Call upgradeRelease: %q", updateRelease)
	//set helm ENV settings
	settings = *helm_settings
	setUpgradeReleaseDefaultValue(updateRelease)
	if len(updateRelease.Chart) == 0 || len(updateRelease.Release) == 0 {
		return nil, errors.New("'upgrade release' requires a release name and a chart name")
	}

	if updateRelease.Install {
		// If a release does not exist, install it. If another error occurs during
		// the check, ignore the error and continue with the upgrade.
		//
		// The returned error is a grpc.rpcError that wraps the message from the original error.
		// So we're stuck doing string matching against the wrapped error, which is nested somewhere
		// inside of the grpc.rpcError message.
		_, err := helmclient.ReleaseHistory(updateRelease.Release, helm.WithMaxHistory(1))
		if err != nil && strings.Contains(err.Error(), driver.ErrReleaseNotFound(updateRelease.Release).Error()) {
			log.Printf("Release %q does not exist. Installing it now.\n", updateRelease.Release)
			installRelease := updateRelease_to_installRelease(updateRelease)
			InstallRelease(helmclient, helm_settings, installRelease)
		}
	}

	chartPath, err := locateChartPath(updateRelease.Chart, updateRelease.Version, updateRelease.Verify, updateRelease.Keyring)
	if err != nil {
		return nil, err
	}
	rawVals := new([]byte)
	// Check chart requirements to make sure all dependencies are present in /charts
	if ch, err := chartutil.Load(chartPath); err == nil {
		if req, err := chartutil.LoadRequirements(ch); err == nil {
			if err := checkDependencies(ch, req); err != nil {
				return nil, err
			}
		}
	}

	rel, err := helmclient.UpdateRelease(
		updateRelease.Release,
		chartPath,
		helm.UpdateValueOverrides(*rawVals),
		helm.UpgradeDryRun(updateRelease.DryRun),
		helm.UpgradeRecreate(updateRelease.Recreate),
		helm.UpgradeDisableHooks(updateRelease.DisableHooks),
		helm.UpgradeTimeout(updateRelease.Timeout),
		helm.ResetValues(updateRelease.ResetValues),
		helm.ReuseValues(updateRelease.ReuseValues),
		helm.UpgradeWait(updateRelease.Wait))
	if err != nil {
		return nil, fmt.Errorf("UPGRADE FAILED: %v", prettyError(err))
	}

	log.Printf("Release %q has been upgraded. Happy Helming!\n", updateRelease.Release)
	// Print the status like status command does
	release := rel.GetRelease()
	if release == nil {
		return nil, nil
	}
	status, err := helmclient.ReleaseStatus(release.Name)
	if err != nil {
		return nil, prettyError(err)
	}
	return status, nil
}

func DeleteRelease(helmclient helm.Interface, deleteRelease *models.DeleteRelease) (*rls.UninstallReleaseResponse, error) {
	log.Printf("Call GetAllReleases: %q", deleteRelease)

	if len(deleteRelease.Name) == 0 {
		return nil, errors.New("'delete release' requires a release name")
	}

	opts := []helm.DeleteOption{
		helm.DeleteDryRun(deleteRelease.DryRun),
		helm.DeleteDisableHooks(deleteRelease.NoHooks),
		helm.DeletePurge(deleteRelease.Purge),
		helm.DeleteTimeout(int64(deleteRelease.Timeout)),
	}
	res, err := helmclient.DeleteRelease(deleteRelease.Name, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func setListReleaseDefaultValue(listRelease *models.ListRelease) {
	if listRelease.Limit == 0 {
		listRelease.Limit = 256
	}
	if len(listRelease.Offset) == 0 {
		listRelease.Offset = ""
	}
	if len(listRelease.Namespace) == 0 {
		listRelease.Namespace = "default"
	}
}

func setRollbackReleaseDefaultValue(updateRelease *models.UpdateRelease) {
	if updateRelease.Timeout == 0 {
		updateRelease.Timeout = 300
	}
}

func setUpgradeReleaseDefaultValue(updateRelease *models.UpdateRelease) {
	if len(updateRelease.Keyring) == 0 {
		updateRelease.Namespace = defaultKeyring()
	}
	if len(updateRelease.Version) == 0 {
		updateRelease.Version = ""
	}
	if len(updateRelease.Namespace) == 0 {
		updateRelease.Namespace = "default"
	}
	if updateRelease.Timeout == 0 {
		updateRelease.Timeout = 300
	}
}

func setInstallReleaseDefaultValue(installRelease *models.InstallReleaseRequest) {
	if len(installRelease.Name) == 0 {
		installRelease.Name = ""
	}
	if len(installRelease.Namespace) == 0 {
		installRelease.Namespace = "default"
	}
	if len(installRelease.Keyring) == 0 {
		installRelease.Keyring = defaultKeyring()
	}
	if installRelease.Timeout == 0 {
		installRelease.Timeout = 300
	}
}

// statusCodes gets the list of status codes that are to be included in the results.
func statusCodes(listRelease *models.ListRelease) []release.Status_Code {
	if listRelease.All {
		return []release.Status_Code{
			release.Status_UNKNOWN,
			release.Status_DEPLOYED,
			release.Status_DELETED,
			release.Status_DELETING,
			release.Status_FAILED,
		}
	}
	status := []release.Status_Code{}
	if listRelease.Deployed {
		status = append(status, release.Status_DEPLOYED)
	}
	if listRelease.Deleted {
		status = append(status, release.Status_DELETED)
	}
	if listRelease.Deleting {
		status = append(status, release.Status_DELETING)
	}
	if listRelease.Failed {
		status = append(status, release.Status_FAILED)
	}
	if listRelease.Superseded {
		status = append(status, release.Status_SUPERSEDED)
	}

	// Default case.
	if len(status) == 0 {
		status = append(status, release.Status_DEPLOYED, release.Status_FAILED)
	}
	return status
}

// prettyError unwraps or rewrites certain errors to make them more user-friendly.
func prettyError(err error) error {
	if err == nil {
		return nil
	}
	// This is ridiculous. Why is 'grpc.rpcError' not exported? The least they
	// could do is throw an interface on the lib that would let us get back
	// the desc. Instead, we have to pass ALL errors through this.
	return errors.New(grpc.ErrorDesc(err))
}


// locateChartPath looks for a chart directory in known places, and returns either the full path or an error.
//
// This does not ensure that the chart is well-formed; only that the requested filename exists.
//
// Order of resolution:
// - current working directory
// - if path is absolute or begins with '.', error out here
// - chart repos in $HELM_HOME
// - URL
//
// If 'verify' is true, this will attempt to also verify the chart.
func locateChartPath(name, version string, verify bool, keyring string) (string, error) {
	name = strings.TrimSpace(name)
	version = strings.TrimSpace(version)
	if fi, err := os.Stat(name); err == nil {
		abs, err := filepath.Abs(name)
		if err != nil {
			return abs, err
		}
		if verify {
			if fi.IsDir() {
				return "", errors.New("cannot verify a directory")
			}
			if _, err := downloader.VerifyChart(abs, keyring); err != nil {
				return "", err
			}
		}
		return abs, nil
	}
	if filepath.IsAbs(name) || strings.HasPrefix(name, ".") {
		return name, fmt.Errorf("path %q not found", name)
	}

	crepo := filepath.Join(settings.Home.Repository(), name)
	if _, err := os.Stat(crepo); err == nil {
		return filepath.Abs(crepo)
	}

	dl := downloader.ChartDownloader{
		HelmHome: settings.Home,
		Out:      os.Stdout,
		Keyring:  keyring,
		Getters:  getter.All(settings),
	}
	if verify {
		dl.Verify = downloader.VerifyAlways
	}

	filename, _, err := dl.DownloadTo(name, version, ".")
	if err == nil {
		lname, err := filepath.Abs(filename)
		if err != nil {
			return filename, err
		}
		//debug("Fetched %s to %s\n", name, filename)
		return lname, nil
	} else if settings.Debug {
		return filename, err
	}

	return filename, fmt.Errorf("file %q not found", name)
}

func generateName(nameTemplate string) (string, error) {
	t, err := template.New("name-template").Funcs(sprig.TxtFuncMap()).Parse(nameTemplate)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	err = t.Execute(&b, nil)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func checkDependencies(ch *chart.Chart, reqs *chartutil.Requirements) error {
	missing := []string{}

	deps := ch.GetDependencies()
	for _, r := range reqs.Dependencies {
		found := false
		for _, d := range deps {
			if d.Metadata.Name == r.Name {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, r.Name)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("found in requirements.yaml, but missing in charts/ directory: %s", strings.Join(missing, ", "))
	}
	return nil
}

// defaultKeyring returns the expanded path to the default keyring.
func defaultKeyring() string {
	return os.ExpandEnv("$HOME/.gnupg/pubring.gpg")
}

func updateRelease_to_installRelease(updateRelease *models.UpdateRelease) *models.InstallReleaseRequest {
	if updateRelease == nil {
		return nil
	}
	installRelease := new(models.InstallReleaseRequest)
	installRelease.Name = updateRelease.Release
	installRelease.Namespace = updateRelease.Namespace
	installRelease.Chart = updateRelease.Chart
	installRelease.DryRun = updateRelease.DryRun
	installRelease.DisableHooks = updateRelease.DisableHooks
	installRelease.Replace = false
	installRelease.NameTemplate = ""
	installRelease.Verify = updateRelease.Verify
	installRelease.Keyring = updateRelease.Keyring
	installRelease.Version = updateRelease.Version
	installRelease.Timeout = updateRelease.Timeout
	installRelease.Wait = updateRelease.Wait

	return installRelease
}
