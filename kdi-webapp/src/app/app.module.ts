import { APP_INITIALIZER, NgModule, CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { HTTP_INTERCEPTORS, HttpClientModule } from '@angular/common/http';
import { MatTableModule } from '@angular/material/table';
import { MatPaginatorModule } from '@angular/material/paginator';
import { MatSortModule } from '@angular/material/sort';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MultiSelectModule } from 'primeng/multiselect';

import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import {
    MsalBroadcastService, MsalModule, MsalService,
    MSAL_INSTANCE
} from '@azure/msal-angular';

import { FileUploadModule } from 'primeng/fileupload';

import { AppRoutingModule } from 'src/app/app-routing.module';
import { MSALInstanceFactory } from 'src/app/auth-config-ms';

import { AuthInterceptor } from 'src/app/auth-interceptor';
import { DateagoPipe } from 'src/app/pipes/date-ago.pipe';

import { AppComponent } from 'src/app/app.component';
import { HomeComponent } from 'src/app/components/home/home.component';
import { LoginComponent } from 'src/app/components/login/login.component';
import { RegisterComponent } from 'src/app/components/register/register.component';
import { SignInMicrosoftComponent } from 'src/app/components/sign-in-microsoft/sign-in-microsoft.component';
import { NavbarComponent } from 'src/app/components/navbar/navbar.component';
import { SidebarComponent } from 'src/app/components/sidebar/sidebar.component';

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
import { EnvironmentDetailsComponent } from 'src/app/components/environments/create-environment/environment-details/environment-details.component';

import { AddMicroserviceYamlComponent } from 'src/app/components/microservices/add-microservice-yaml/add-microservice-yaml.component';
import { MicroserviceDetailsComponent } from 'src/app/components/microservices/microservice-details/microservice-details.component';

import { MessageModule } from 'primeng/message';
import { MessagesModule } from 'primeng/messages';
import { MessageService } from 'primeng/api';
import { ToastModule } from 'primeng/toast';
import { AvatarModule } from 'primeng/avatar';
import { BadgeModule } from 'primeng/badge';
import { ProgressSpinnerModule } from 'primeng/progressspinner';
import { RippleModule } from 'primeng/ripple';
import { DropdownModule } from 'primeng/dropdown';
import { InputTextModule } from 'primeng/inputtext';

import { ErrorCatchingInterceptor } from 'src/app/error-catching.interceptor';
import { CacheInterceptor } from 'src/app/cache.interceptor';
import { AddOpenshiftClusterComponent } from 'src/app/components/clusters/add-cluster/opensfhit-cluster/openshift-cluster.component';
import { AddEksClusterComponent } from 'src/app/components/clusters/add-cluster/eks-cluster/eks-cluster.component';

export function initializeMsal(msalService: MsalService): () => Promise<void> {
    return () => msalService.instance.initialize();
}

@NgModule({
    declarations: [
        DateagoPipe,
        AppComponent,
        LoginComponent,
        RegisterComponent,
        SignInMicrosoftComponent,
        HomeComponent,
        NavbarComponent,
        SidebarComponent,
        CreateTeamspaceComponent,
        AddClusterComponent,
        AddEksClusterComponent,
        AddLocalClusterComponent,
        AddOpenshiftClusterComponent,
        CreateProjectComponent,
        ListProjectsComponent,
        ListClustersComponent,
        ListTeamspacesComponent,
        ViewTeamspaceComponent,
        DeploymentWithHelmComponent,
        CreateEnvironmentComponent,
        ProjectDetailsComponent,
        AddMicroserviceYamlComponent,
        ProjectDetailsComponent,
        EnvironmentDetailsComponent,
        MicroserviceDetailsComponent
    ],
    imports: [
        ReactiveFormsModule,
        BrowserModule,
        BrowserAnimationsModule,
        BrowserAnimationsModule,
        HttpClientModule,
        AppRoutingModule,
        MsalModule,
        MatTableModule,
        MatPaginatorModule,
        MatSortModule,
        MatInputModule,
        MatSelectModule,
        MultiSelectModule,
        FormsModule,
        MessageModule,
        MessagesModule,
        FileUploadModule,
        ToastModule,
        AvatarModule,
        BadgeModule,
        ProgressSpinnerModule,
        RippleModule,
        MessageModule,
        DropdownModule,
        InputTextModule,
    ],
    schemas: [
        CUSTOM_ELEMENTS_SCHEMA
    ],
    exports: [
        NavbarComponent,
        SidebarComponent,
    ],
    providers: [
        {
            provide: HTTP_INTERCEPTORS,
            useClass: AuthInterceptor,
            multi: true
        },
        {
            provide: HTTP_INTERCEPTORS,
            useClass: ErrorCatchingInterceptor,
            multi: true
        },
        {
            provide: HTTP_INTERCEPTORS,
            useClass: CacheInterceptor,
            multi: true
        },
        {
            provide: MSAL_INSTANCE,
            useFactory: MSALInstanceFactory
        },
        {
            provide: APP_INITIALIZER,
            useFactory: initializeMsal,
            deps: [MsalService],
            multi: true
        },
        MsalService,
        MsalBroadcastService,
        MessageService,
    ],
    bootstrap: [AppComponent,]
})
export class AppModule { }
