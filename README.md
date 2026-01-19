# Kubernetes Deployment Interface (KDI)
In this repo, we've created an api in order to make deployment on a k8s cluster locally.

The monolyth version (with the contributions of KDI team members) can be found here : https://github.com/kuro-jojo/kdi
## Prerequisites
- Go [https://go.dev/doc/install](https://go.dev/doc/install)
- K8s cluster

## Run 
To start the api server 
1. Clone this repo
2. Install the dependencies
```bash
  cd kdi/api
  go mod tidy
```
3. Run the api server
```bash
  cd kdi/api
  go run . 
```

You can then open your browser and go to [http://localhost:8080](http://localhost:8080).

## Usage
There are two parts in the application:
- The first part is the frontend
- The second part is the backend (api)

### Backend
The backend is an api server which comprises two services:
- The first service is related to the communication with the k8s cluster (**kubernetes service**)
- The second service is related to the communication with the frontend, the management of users and projects and thus the communication with the database. (**web service**)

#### Kubernetes service
The kubernetes service is the service that will communicate with the k8s cluster. It will handle the authentication and the creation of the resources.

##### Prerequisites

In order to be authenticate and authorized in the cluster, the application need some rights :

- So first, the cluster admin (you) has to create a service account for our application
- Then add the required roles for all the namespaces the application will access. 
**The namespace must be create before creating those objects**
- Create a token for the service account. There are multiple ways. 
One is to run `kubectl create token service-account-name` and copy the generated token in the token field.
This token will expire in 10min (default time) 
` kubectl create token service-account-name --duration 3600s` to custom the expiration date.


##### Endpoints

    /kubernetes/resources/{resourceType} : to create a resource
      - /deployment : to create a deployment
      - /service : to create a service  


## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.
