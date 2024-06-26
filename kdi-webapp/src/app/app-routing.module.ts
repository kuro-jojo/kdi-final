import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { HomeComponent } from 'src/app/components/home/home.component';
import { BrowserUtils } from '@azure/msal-browser';


import { AuthGuard } from 'src/app/auth.guard';
import { LoginComponent } from 'src/app/components/login/login.component';
import { RegisterComponent } from 'src/app/components/register/register.component';

import { AddClusterComponent } from 'src/app/components/clusters/add-cluster/add-cluster.component';
import { ListClustersComponent } from 'src/app/components/clusters/list-clusters/list-clusters.component';
import { AddLocalClusterComponent } from 'src/app/components/clusters/add-cluster/local-cluster/local-cluster.component';

import { CreateProjectComponent } from 'src/app/components/projects/create-project/create-project.component';
import { CreateTeamspaceComponent } from 'src/app/components/teamspaces/create-teamspace/create-teamspace.component';
import { ListTeamspacesComponent } from 'src/app/components/teamspaces/list-teamspaces/list-teamspaces.component';
import { ViewTeamspaceComponent } from 'src/app/components/teamspaces/view-teamspace/view-teamspace.component';

import { ProjectDetailsComponent } from 'src/app/components/projects/project-details/project-details.component';
import { ListProjectsComponent } from 'src/app/components/projects/list-projects/list-projects.component';
import { DeploymentWithHelmComponent } from 'src/app/components/deployment-with-helm/deployment-with-helm.component';
import { CreateEnvironmentComponent } from 'src/app/components/environments/create-environment/create-environment.component';
import { AddMicroserviceYamlComponent } from 'src/app/components/microservices/add-microservice-yaml/add-microservice-yaml.component';
import { ServerDownComponent } from 'src/app/components/errors/server-down/server-down.component';
import { PageNotFoundComponent } from 'src/app/components/errors/page-not-found/page-not-found.component';
import { EnvironmentDetailsComponent } from 'src/app/components/environments/create-environment/environment-details/environment-details.component';
import { MicroserviceDetailsComponent } from 'src/app/components/microservices/microservice-details/microservice-details.component';
import { AddOpenshiftClusterComponent } from 'src/app/components/clusters/add-cluster/opensfhit-cluster/openshift-cluster.component';
import { AddEksClusterComponent } from 'src/app/components/clusters/add-cluster/eks-cluster/eks-cluster.component';


const routes: Routes = [
    { path: "", component: HomeComponent, canActivate: [AuthGuard] },
    { path: "login", component: LoginComponent },
    { path: "register", component: RegisterComponent },

    { path: "projects/add", component: CreateProjectComponent, title: "Add new project", canActivate: [AuthGuard] },
    { path: "projects/:projectId", component: ProjectDetailsComponent, canActivate: [AuthGuard], title: "Project details" },
    { path: "projects/:id/environments/add", component: CreateEnvironmentComponent, title: "Add new environment", canActivate: [AuthGuard] },
    { path: "projects", component: ListProjectsComponent, canActivate: [AuthGuard], title: "List of projects" },

    { path: "environments/:id/deployments/with-yaml", component: AddMicroserviceYamlComponent, title: "Deploy with helm", canActivate: [AuthGuard] },
    { path: "environments/add", component: CreateEnvironmentComponent, title: "Add new environment", canActivate: [AuthGuard] },
    { path: "environments/:envId", component: EnvironmentDetailsComponent, canActivate: [AuthGuard], title: "Environment details" },
    { path: "environments/:envId/microservices/:mId", component: MicroserviceDetailsComponent, canActivate: [AuthGuard], title: "Microservice details" },

    { path: "clusters/local/:id/edit", component: AddLocalClusterComponent, canActivate: [AuthGuard], title: "Edit cluster" },
    { path: "clusters/eks/:id/edit", component: AddEksClusterComponent, canActivate: [AuthGuard], title: "Edit cluster" },
    { path: "clusters/openshift/:id/edit", component: AddOpenshiftClusterComponent, canActivate: [AuthGuard], title: "Edit cluster" },
    { path: "clusters/add/eks", component: AddEksClusterComponent, canActivate: [AuthGuard], title: "Add new eks cluster" },
    { path: "clusters/add/local", component: AddLocalClusterComponent, canActivate: [AuthGuard], title: "Add new local cluster" },
    { path: "clusters/add/openshift", component: AddOpenshiftClusterComponent, canActivate: [AuthGuard], title: "Add new openshift cluster" },
    { path: "clusters/add", component: AddClusterComponent, canActivate: [AuthGuard], title: "Add new cluster" },
    { path: "clusters", component: ListClustersComponent, canActivate: [AuthGuard], title: "List of clusters" },

    { path: "teamspaces/add", component: CreateTeamspaceComponent, canActivate: [AuthGuard], title: "Add new teamspace" },
    { path: "teamspaces/:teamId", component: ViewTeamspaceComponent, canActivate: [AuthGuard], title: "Team details" },
    { path: "teamspaces", component: ListTeamspacesComponent, canActivate: [AuthGuard], title: "List of teamspaces" },

    { path: "deployments/helm", component: DeploymentWithHelmComponent, canActivate: [AuthGuard], title: "Deploy with helm" },

    { path: "500", component: ServerDownComponent, title: "Server down" },
    { path: "**", component: PageNotFoundComponent, title: "Page not found" }
];

@NgModule({
    imports: [RouterModule.forRoot(routes, {
        // Don't perform initial navigation in iframes or popups
        initialNavigation:
            !BrowserUtils.isInIframe() && !BrowserUtils.isInPopup()
                ? "enabledNonBlocking"
                : "disabled", // Set to enabledBlocking to use Angular Universal
    })],
    exports: [RouterModule]
})
export class AppRoutingModule { }
