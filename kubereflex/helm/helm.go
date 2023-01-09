package helm

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/flock"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
	"helm.sh/helm/v3/pkg/strvals"
	"sigs.k8s.io/yaml"
)

var settings *cli.EnvSettings = cli.New()

func setSettings(namespace string, kubeconfig *string) {
	os.Setenv("HELM_NAMESPACE", namespace)
	settings.SetNamespace(namespace)
	settings.KubeConfig = *kubeconfig
}

// Install set helm settings up, perform repository updates and install the chart which is specified
func Install(repositoryName string, chartName string, releaseName string, namespace string, args map[string]string, kubeconfig *string) {
	setSettings(namespace, kubeconfig)
	RepositoryUpdate()
	installChart(releaseName, repositoryName, chartName, args)
}

// Uninstall set helm settings up and uninstall the chart which is specified
func Uninstall(releaseName string, namespace string, kubeconfig *string) {
	setSettings(namespace, kubeconfig)
	uninstallChart(releaseName)
}

// IsRepositoryExists check if given repositoryName already exists in repo.File
func IsRepositoryExists(repositoryName string) bool {
	var repoFile = readRepositoryFile(settings.RepositoryConfig)

	if repoFile.Has(repositoryName) {
		fmt.Printf("Nice! %s already in the repos!\n", repositoryName)
		return true
	}

	return false
}

// RepositoryAdd adds helm repository to current helm instance
func RepositoryAdd(repositoryName, chartUrl string) {
	var repoFile = readRepositoryFile(settings.RepositoryConfig)

	newChart := repo.Entry{
		Name: repositoryName,
		URL:  chartUrl,
	}

	repository, err := repo.NewChartRepository(&newChart, getter.All(settings))
	if err != nil {
		log.Fatal(err)
	}

	if _, err := repository.DownloadIndexFile(); err != nil {
		err := errors.Wrapf(err, "Ouch, looks like %q is not a valid chart repository or cannot be reached\n", chartUrl)
		log.Fatal(err)
	}

	repoFile.Update(&newChart)

	if err := repoFile.WriteFile(settings.RepositoryConfig, 0644); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Great! %q has been added to your repositories\n", settings.RepositoryConfig)
}

// RepositoryUpdate updates charts for all helm repos
func RepositoryUpdate() {
	var repoFile = readRepositoryFile(settings.RepositoryConfig)

	var repos []*repo.ChartRepository
	for _, cfg := range repoFile.Repositories {
		repository, err := repo.NewChartRepository(cfg, getter.All(settings))
		if err != nil {
			log.Fatal(err)
		}
		repos = append(repos, repository)
	}

	fmt.Println("Hang tight while we grab the latest from your chart repositories...")
	var wg sync.WaitGroup
	for _, repository := range repos {
		wg.Add(1)
		go func(repository *repo.ChartRepository) {
			defer wg.Done()
			if _, err := repository.DownloadIndexFile(); err != nil {
				fmt.Printf("Sad. Unable to get an update from the %q chart repository (%s):\n\t%s\n", repository.Config.Name, repository.Config.URL, err)
			} else {
				fmt.Printf("Yay! Successfully got an update from the %q chart repository\n", repository.Config.Name)
			}
		}(repository)
	}
	wg.Wait()
	fmt.Println("Alright! Update Complete. ⎈ Happy Helming! ⎈")
}

// installChart perform a chart install
func installChart(releaseName, repositoryName, chartName string, args map[string]string) {
	fmt.Printf("Install %s chart from %s repository...\n", chartName, repositoryName)
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), debug); err != nil {
		log.Fatal(err)
	}
	client := action.NewInstall(actionConfig)

	if client.Version == "" && client.Devel {
		client.Version = ">0.0.0-0"
	}

	client.ReleaseName = releaseName
	chartPath, err := client.ChartPathOptions.LocateChart(fmt.Sprintf("%s/%s", repositoryName, chartName), settings)
	if err != nil {
		log.Fatal(err)
	}

	getter.All(settings)

	p := getter.All(settings)
	valueOpts := &values.Options{}
	vals, err := valueOpts.MergeValues(p)
	if err != nil {
		log.Fatal(err)
	}

	if err := strvals.ParseInto(args["set"], vals); err != nil {
		log.Fatal(errors.Wrap(err, "Nah, failed parsing --set data\n"))
	}

	chartRequested, err := loader.Load(chartPath)
	if err != nil {
		log.Fatal(err)
	}

	validInstallableChart, err := isChartInstallable(chartRequested)
	if !validInstallableChart {
		log.Fatal(err)
	}

	if req := chartRequested.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(chartRequested, req); err != nil {
			if client.DependencyUpdate {
				manager := &downloader.Manager{
					Out:              os.Stdout,
					ChartPath:        chartPath,
					Keyring:          client.ChartPathOptions.Keyring,
					SkipUpdate:       false,
					Getters:          p,
					RepositoryConfig: settings.RepositoryConfig,
					RepositoryCache:  settings.RepositoryCache,
				}
				if err := manager.Update(); err != nil {
					log.Fatal(err)
				}
			} else {
				log.Fatal(err)
			}
		}
	}

	client.Namespace = settings.Namespace()
	release, err := client.Run(chartRequested, vals)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s is deployed\n", release.Name)
}

// uninstallChart perform a chart uninstall
func uninstallChart(releaseName string) {
	fmt.Printf("Uninstall %s chart\n", releaseName)
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), debug); err != nil {
		log.Fatal(err)
	}
	client := action.NewUninstall(actionConfig)

	release, err := client.Run(releaseName)
	if err != nil {
		fmt.Printf("%s\n", err)
	} else {
		fmt.Printf("%s is uninstalled\n", release.Release.Name)		
	}
}

// isChartInstallable check chart type is installable
func isChartInstallable(chart *chart.Chart) (bool, error) {
	switch chart.Metadata.Type {
	case "", "application":
		return true, nil
	}

	return false, errors.Errorf("%s charts are not installable!\n", chart.Metadata.Type)
}

// readRepositoryFile read repository file and return with that
func readRepositoryFile(repositoryFile string) repo.File {
	err := os.MkdirAll(filepath.Dir(repositoryFile), os.ModePerm)
	if err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}

	fileLock := flock.New(strings.Replace(repositoryFile, filepath.Ext(repositoryFile), ".lock", 1))
	lockCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	locked, err := fileLock.TryLockContext(lockCtx, time.Second)
	if err == nil && locked {
		defer fileLock.Unlock()
	}
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.ReadFile(repositoryFile)
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	var repoFile repo.File
	if err := yaml.Unmarshal(file, &repoFile); err != nil {
		log.Fatal(err)
	}

	return repoFile
}

func debug(format string, v ...interface{}) {
	// TODO: Output only in debug mode
	// format = fmt.Sprintf("[debug] %s\n", format)
	// log.Output(2, fmt.Sprintf(format, v...))
}
