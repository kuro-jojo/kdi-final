package controllers

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v2"

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
)

func RepoAdd(name, url string) {

	settings := cli.New()

	//Récupération du fichier de configuration du référentiel
	repoFile := settings.RepositoryConfig

	//Création du répertoire du fichier de configuration si nécessaire
	err := os.MkdirAll(filepath.Dir(repoFile), os.ModePerm)
	if err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}

	// Cela crée un verrou de fichier pour synchroniser les processus.
	//Il permet d'éviter les conflits lorsque plusieurs processus modifient le fichier de configuration du référentiel simultanément.
	fileLock := flock.New(strings.Replace(repoFile, filepath.Ext(repoFile), ".lock", 1))
	lockCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	locked, err := fileLock.TryLockContext(lockCtx, time.Second)
	if err == nil && locked {
		defer fileLock.Unlock()
	}
	if err != nil {
		log.Fatal(err)
	}

	//Lecture du fichier de configuration du référentiel
	b, err := os.ReadFile(repoFile)
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	//désérialise le contenu du fichier YAML dans un objet repo.File, qui contient les entrées des référentiels Helm.
	var f repo.File
	if err := yaml.Unmarshal(b, &f); err != nil {
		log.Fatal(err)
	}

	//Vérification de l'existence du référentiel
	if f.Has(name) {
		fmt.Printf("repository name (%s) already exists\n", name)
		return
	}

	//Création d'un nouvel objet de référentiel
	c := repo.Entry{
		Name: name,
		URL:  url,
	}

	r, err := repo.NewChartRepository(&c, getter.All(settings))
	if err != nil {
		log.Fatal(err)
	}

	//Téléchargement du fichier d'index du référentiel
	if _, err := r.DownloadIndexFile(); err != nil {
		err := errors.Wrapf(err, "looks like %q is not a valid chart repository or cannot be reached", url)
		log.Fatal(err)
	}

	//Mise à jour du fichier de configuration du référentiel
	f.Update(&c)

	if err := f.WriteFile(repoFile, 0644); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%q has been added to your repositories\n", name)
}

// RepoUpdate updates charts for all helm repos
func RepoUpdate() {

	settings := cli.New()

	//Récupération du fichier de configuration du référentiel
	repoFile := settings.RepositoryConfig
	f, err := repo.LoadFile(repoFile)

	//Vérification de l'existence de référentiels dans le fichier de configuration
	if os.IsNotExist(errors.Cause(err)) || len(f.Repositories) == 0 {
		log.Fatal(errors.New("no repositories found. You must add one before updating"))
	}

	//Création des objets de référentiel à partir du fichier de configuration
	var repos []*repo.ChartRepository
	for _, cfg := range f.Repositories {
		r, err := repo.NewChartRepository(cfg, getter.All(settings))
		if err != nil {
			log.Fatal(err)
		}
		repos = append(repos, r)
	}
	fmt.Printf("Hang tight while we grab the latest from your chart repositories...\n")

	//Téléchargement des fichiers d'index de chaque référentiel en parallèle
	var wg sync.WaitGroup
	for _, re := range repos {
		wg.Add(1)
		go func(re *repo.ChartRepository) {
			defer wg.Done()
			if _, err := re.DownloadIndexFile(); err != nil {
				fmt.Printf("...Unable to get an update from the %q chart repository (%s):\n\t%s\n", re.Config.Name, re.Config.URL, err)
			} else {
				fmt.Printf("...Successfully got an update from the %q chart repository\n", re.Config.Name)
			}
		}(re)
	}
	wg.Wait()
	fmt.Printf("Update Complete. ⎈ Happy Helming!⎈\n")

}

// InstallChart
func InstallChartFromRepo(name, repo, chart string, args map[string]string) {

	settings := cli.New()
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("KDI_KDI_HELM_DRIVER"), log.Printf); err != nil {
		log.Fatal(err)
	}
	client := action.NewInstall(actionConfig)

	if client.Version == "" && client.Devel {
		client.Version = ">0.0.0-0"
	}
	//Définition des options de l'installation du chart
	client.ReleaseName = name
	cp, err := client.ChartPathOptions.LocateChart(fmt.Sprintf("%s/%s", repo, chart), settings)
	if err != nil {
		log.Fatal(err)
	}

	//debug("CHART PATH: %s\n", cp)

	p := getter.All(settings)
	valueOpts := &values.Options{}
	vals, err := valueOpts.MergeValues(p)
	if err != nil {
		log.Fatal(err)
	}

	//Analyse et ajout des valeurs définies par l'utilisateur
	if err := strvals.ParseInto(args["set"], vals); err != nil {
		log.Fatal(errors.Wrap(err, "failed parsing --set data"))
	}

	chartRequested, err := loader.Load(cp)
	if err != nil {
		log.Fatal(err)
	}

	//vérifie la validité du chart
	validInstallableChart, err := isChartInstallable(chartRequested)
	if !validInstallableChart {
		log.Fatal(err)
	}

	//vérifie si le chart a des dépendances et si elles sont satisfaites
	if req := chartRequested.Metadata.Dependencies; req != nil {
		// If CheckDependencies returns an error, we have unfulfilled dependencies.
		// https://github.com/helm/helm/issues/2209
		if err := action.CheckDependencies(chartRequested, req); err != nil {
			if client.DependencyUpdate {
				man := &downloader.Manager{
					Out:              os.Stdout,
					ChartPath:        cp,
					Keyring:          client.ChartPathOptions.Keyring,
					SkipUpdate:       false,
					Getters:          p,
					RepositoryConfig: settings.RepositoryConfig,
					RepositoryCache:  settings.RepositoryCache,
				}
				if err := man.Update(); err != nil {
					log.Fatal(err)
				}
			} else {
				log.Fatal(err)
			}
		}
	}

	//client.Namespace = settings.Namespace()
	release, err := client.Run(chartRequested, vals)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(release.Manifest)
}

func isChartInstallable(ch *chart.Chart) (bool, error) {
	switch ch.Metadata.Type {
	case "", "application":
		return true, nil
	}
	return false, errors.Errorf("%s charts are not installable", ch.Metadata.Type)
}
