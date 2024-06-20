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

import { AppRoutingModule } from './app-routing.module';
import { MSALInstanceFactory } from './auth-config-ms';

import { AuthInterceptor } from './auth-interceptor';
import { DateagoPipe } from './pipes/date-ago.pipe';

import { AppComponent } from './app.component';
import { HomeComponent } from 'src/app/components/home/home.component';
import { LoginComponent } from 'src/app/components/login/login.component';
import { RegisterComponent } from 'src/app/components/register/register.component';
import { SignInMicrosoftComponent } from './components/sign-in-microsoft/sign-in-microsoft.component';
import { NavbarComponent } from './components/navbar/navbar.component';
import { SidebarComponent } from './components/sidebar/sidebar.component';

import { AddClusterComponent } from './components/clusters/add-cluster/add-cluster.component';
import { ListClustersComponent } from './components/clusters/list-clusters/list-clusters.component';
import { AddLocalClusterComponent } from './components/clusters/add-cluster/local-cluster/local-cluster.component';

import { CreateProjectComponent } from './components/projects/create-project/create-project.component';
import { CreateTeamspaceComponent } from './components/teamspaces/create-teamspace/create-teamspace.component';
import { ListTeamspacesComponent } from './components/teamspaces/list-teamspaces/list-teamspaces.component';
import { ViewTeamspaceComponent } from './components/teamspaces/view-teamspace/view-teamspace.component';

import { ProjectDetailsComponent } from './components/projects/project-details/project-details.component';
import { ListProjectsComponent } from './components/projects/list-projects/list-projects.component';
import { DeploymentWithHelmComponent } from './components/deployment-with-helm/deployment-with-helm.component';
import { CreateEnvironmentComponent } from './components/environments/create-environment/create-environment.component';
import { EnvironmentDetailsComponent } from './components/environments/create-environment/environment-details/environment-details.component';

import { AddMicroserviceYamlComponent } from './components/microservices/add-microservice-yaml/add-microservice-yaml.component';

import { MessageModule } from 'primeng/message';
import { MessagesModule } from 'primeng/messages';
import { MessageService } from 'primeng/api';
import { ToastModule } from 'primeng/toast';
import { AvatarModule } from 'primeng/avatar';
import { BadgeModule } from 'primeng/badge';
import { ProgressSpinnerModule } from 'primeng/progressspinner';
import { RippleModule } from 'primeng/ripple';
import { ErrorCatchingInterceptor } from './error-catching.interceptor';
import { CacheInterceptor } from './cache.interceptor';
import { MicroserviceDetailsComponent } from './components/microservices/microservice-details/microservice-details.component';

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
        AddLocalClusterComponent,
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
