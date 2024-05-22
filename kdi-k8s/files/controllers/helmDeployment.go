package controllers

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kuro-jojo/kdi-k8s/models"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

func HelmDeploymentFromRepo(c *gin.Context) {

	var repoEntry models.RepoEntry

	// Récupération de la valeur de "set" du formulaire POST
	setValues := c.PostForm("set")

	// Création d'un map pour stocker les paires clé-valeur
	args := make(map[string]string)

	// Division de la chaîne de valeurs en paires clé-valeur
	setPairs := strings.Split(setValues, ",")
	for _, pair := range setPairs {
		keyValue := strings.Split(pair, "=")
		if len(keyValue) == 2 {
			args[keyValue[0]] = keyValue[1]
		}
	}

	// Bind JSON fourni par l'utilisateur
	if err := c.ShouldBindJSON(&repoEntry); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	os.Setenv("HELM_NAMESPACE", repoEntry.Namespace)

	// Add helm repo
	RepoAdd(repoEntry.RepoName, repoEntry.RepoUrl)
	// Update charts from the helm repo
	RepoUpdate()
	// Install charts
	InstallChartFromRepo(repoEntry.ReleaseName, repoEntry.RepoName, repoEntry.ChartName, args)

}

func GetReposList(c *gin.Context) {
	settings := cli.New()

	//Récupération du fichier de configuration du référentiel
	repoFile := settings.RepositoryConfig
	f, err := repo.LoadFile(repoFile)

	//Vérification de l'existence des repos dans le fichier de configuration
	if os.IsNotExist(errors.Cause(err)) || len(f.Repositories) == 0 {
		log.Fatal(errors.New("no repositories found. You must add one before updating"))
	}

	var repos []*repo.ChartRepository
	for _, cfg := range f.Repositories {
		r, err := repo.NewChartRepository(cfg, getter.All(settings))
		if err != nil {
			log.Fatal(err)
		}
		repos = append(repos, r)
	}
	// Affichage des noms des référentiels
	fmt.Println("List of repositories:")
	for _, r := range repos {
		fmt.Printf("- %s\n", r.Config.Name)
	}

	c.JSON(http.StatusOK, gin.H{"repositories": repos})
}

/*func ListCharts(c *gin.Context) {
	settings := cli.New()

	// Récupération du fichier de configuration du référentiel
	repoFile := settings.RepositoryConfig
	f, err := repo.LoadFile(repoFile)
	if os.IsNotExist(errors.Cause(err)) || len(f.Repositories) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no repositories found"})
		return
	}

	var reposWithCharts []map[string]interface{}
	for _, cfg := range f.Repositories {
		// Initialiser le référentiel
		r, err := repo.NewChartRepository(cfg, getter.All(settings))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Télécharger l'index des charts
		if _, err := r.DownloadIndexFile(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Récupérer l'index des charts
		index := r.Index()
		/*if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var charts []string
		for chartName := range index.Entries {
			charts = append(charts, chartName)
		}

		// Ajouter les informations sur les charts disponibles dans la réponse
		reposWithCharts = append(reposWithCharts, map[string]interface{}{
			"repository": cfg.Name,
			"charts":     charts,
		})
	}

	c.JSON(http.StatusOK, reposWithCharts)
}*/

func HelmDeployment(c *gin.Context) {

	//var repoEntry RepoEntry

	// Bind JSON fourni par l'utilisateur
	/*if err := c.ShouldBindJSON(&namespace); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := c.ShouldBindJSON(&releaseName); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}*/

	namespace, _ := c.GetPostForm("namespace")
	releaseName, _ := c.GetPostForm("releaseName")

	os.Setenv("HELM_NAMESPACE", namespace)

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.String(400, "Bad Request - No file provided")
		fmt.Printf("Problème ici")
		return
	} else {
		fmt.Printf("successfully uploaded")
	}
	defer file.Close()

	// Read the file content
	chartContent, err := io.ReadAll(file)
	if err != nil {
		c.String(500, "Internal Server Error - Unable to read file content")
		return
	}

	// Deploy the chart to Kubernetes cluster
	err = deployChart(chartContent, releaseName)
	if err != nil {
		c.String(500, "Internal Server Error - Unable to deploy chart")
		return
	}

	//c.String(200, "Chart deployed successfully")
}

func deployChart(chartContent []byte, releaseName string) error {
	// Initialize Helm action configuration
	settings := cli.New()

	// Create a Helm action configuration
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("KDI_HELM_DRIVER"), log.Printf); err != nil {
		return err
	}

	// Initialize Helm install action
	install := action.NewInstall(actionConfig)

	// Set namespace and release name
	//install.Namespace = namespace
	install.ReleaseName = releaseName
	/*fmt.Printf("release Name ", releaseName)
	fmt.Printf("Namespace", namespace)*/

	// Create an io.Reader from the byte slice
	chartReader := bytes.NewReader(chartContent)

	// Load the chart content
	chart, err := loader.LoadArchive(chartReader)
	if err != nil {
		return err
	}

	// Run Helm install action
	_, err = install.Run(chart, nil)
	if err != nil {
		fmt.Println("Erreur lors de l'installation du chart Helm:", err)
		return err
	}

	return nil

}

/*func SearchChart(c *gin.Context) {
	var keyword string

	/*if err := c.ShouldBindJSON(&keyword); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	//SearchAction(keyword)
}*/
