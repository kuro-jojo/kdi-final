import { Component, ViewChild } from '@angular/core';
import { ToastComponent } from 'src/app/components/toast/toast.component';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Cluster } from 'src/app/_interfaces/cluster';
import { ActivatedRoute, Router } from '@angular/router';
import { ClusterService } from 'src/app/_services/cluster.service';
import { first } from 'rxjs';
import { HttpErrorResponse } from '@angular/common/http';
import { EnvironmentService } from 'src/app/_services/environment.service';
import { UserService } from 'src/app/_services';
import { Project } from 'src/app/_interfaces/project';
import { ProjectService } from 'src/app/_services/project.service';
import { Environment } from 'src/app/_interfaces/environment';
import { ServerService } from 'src/app/_services/server.service';

@Component({
    selector: 'app-create-environment',
    templateUrl: './create-environment.component.html',
    styleUrl: './create-environment.component.css'
})
export class CreateEnvironmentComponent {

    @ViewChild(ToastComponent) toastComponent!: ToastComponent;
    environmentForm: FormGroup;
    submitted = false;
    clusters: { clusters: Cluster[], size: number } = { clusters: [], size: 0 };
    projects: { projects: Project[], size: number } = { projects: [], size: 0 };
    projectID: string = '';
    environment!: Environment;

    constructor(
        private formBuilder: FormBuilder,
        private router: Router,
        private route: ActivatedRoute,
        private userService: UserService,
        private clusterService: ClusterService,
        private projectService: ProjectService,
        private serverService: ServerService,
        private environmentService: EnvironmentService,) {
        this.environmentForm = new FormGroup({});
    }

    ngOnInit() {
        this.environment = {
            Name: '',
            ClusterID: '',
            ProjectID: ''
        };
        this.environmentForm = this.formBuilder.group({
            Name: ['', Validators.required],
            Description: ['', Validators.minLength(6)],
            ClusterID: ['', Validators.required],
            ProjectID: ['', Validators.required],
        });

        this.projectID = this.route.snapshot.params['id'];
        if (this.projectID) {
            this.environment.ProjectID = this.projectID;
            this.environmentForm.controls['ProjectID'].setValue(this.projectID);
            this.environmentForm.controls['ProjectID'].disable();
        }

        this.serverService.serverStatus().subscribe({
            next: (resp) => {
                this.clusterService.getClusters().subscribe(
                    (resp) => { this.clusters = resp; }
                )
                this.projectService.listProjects().subscribe(
                    (resp) => { this.projects = resp; }
                )
            }, error: () => {
            }
        });


    }
    get formControls() { return this.environmentForm.controls; }

    onSubmit() {
        this.submitted = true;
        // stop here if form is invalide
        if (this.environmentForm.invalid) {
            return;
        }
        this.environment = { ...this.environment, ...this.environmentForm.value };
        if (this.userService.isAuthentificated) {
            this.environmentService.createEnvironment(this.environment)
                .pipe(first())
                .subscribe({
                    next: (resp) => {
                        this.toastComponent.message = "You have successfully created an environment!";
                        this.toastComponent.toastType = 'success';
                        this.triggerToast();
                        this.router.navigate(['/projects/' + this.environmentForm.controls['ProjectID'].value])
                    },
                    error: (error: HttpErrorResponse) => {
                        this.toastComponent.message = error.error.message;
                        this.toastComponent.toastType = 'danger';
                        if (error.status == 0) {
                            this.toastComponent.message = "Server is not available";
                            this.toastComponent.toastType = 'info';
                        }
                        this.triggerToast();
                    },
                    complete: () => {
                        console.log("Environment created successfully");
                    }
                })
        } else {
            this.toastComponent.message = 'Token invalide';
            this.toastComponent.toastType = 'danger';
        }
    }

    triggerToast(): void {
        this.toastComponent.showToast();
    }

}
