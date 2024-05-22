package files

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kuro-jojo/kdi-k8s/models"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

// ProcessUploadedFile processes the uploaded file and returns a list of kubernetes objects
func ProcessUploadedFile(c *gin.Context, file *multipart.FileHeader) ([]models.KubeObject, int, string) {
	log.Printf("Processing file %s", file.Filename)
	filename := filepath.Base(file.Filename)
	if !IsYAML(filename) {
		log.Printf("Only YAML files are supported")
		return nil, http.StatusBadRequest, filename + " : Only YAML files are supported."
	}

	uploadedFile, err := file.Open()
	if err != nil {
		log.Printf("Error opening the file %v", err)
		return nil, http.StatusInternalServerError, filename + " : " + err.Error()
	}
	defer uploadedFile.Close()

	content, err := io.ReadAll(uploadedFile)
	if err != nil {
		log.Printf("Error reading the file %v", err)
		return nil, http.StatusInternalServerError, filename + " : " + err.Error()
	}

	objs, err := getKuberbenetesObjectFromFile(string(content))
	if err != nil {
		log.Printf("Error getting kubernetes objects from file %s : %s", file.Filename, err.Error())
		return nil, http.StatusInternalServerError, filename + " : " + err.Error()
	}
	return objs, 0, ""
}

// return a list of kubernetes objects from a file
func getKuberbenetesObjectFromFile(fileContent string) ([]models.KubeObject, error) {
	log.Printf("Processing file content")
	sections := strings.Split(string(fileContent), "\n---\n")
	decode := scheme.Codecs.UniversalDeserializer().Decode
	objects := make([]models.KubeObject, 0)
	// Print each section
	for _, section := range sections {
		obj, _, err := decode([]byte(section), nil, nil)
		if err != nil {
			fmt.Printf("Error decoding the object: %v\n", err)
			return nil, err
		}
		switch obj.GetObjectKind().GroupVersionKind().Kind {
		case "Deployment":
			objects = append(objects, &models.Deployment{Deployment: obj.(*appv1.Deployment)})
		case "Service":
			objects = append(objects, &models.Service{Service: obj.(*corev1.Service)})

		// TODO : Add other kubernetes objects here
		default:
			log.Printf("Object kind: %s not supported\n", obj.GetObjectKind().GroupVersionKind().Kind)
		}
	}
	return objects, nil
}

func IsYAML(filename string) bool {
	// Convert the filename to lowercase to handle cases like ".YAML" or ".Yml"
	filename = strings.ToLower(filename)
	return strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml")
}
