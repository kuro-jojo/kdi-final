import { Component } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Cluster } from 'src/app/_interfaces/cluster';
import { ActivatedRoute, Router } from '@angular/router';
import { ClusterService } from 'src/app/_services/cluster.service';
import { first } from 'rxjs';
import { HttpErrorResponse } from '@angular/common/http';
import { EnvironmentService } from 'src/app/_services/environment.service';
import { Project } from 'src/app/_interfaces/project';
import { ProjectService } from 'src/app/_services/project.service';
import { Environment } from 'src/app/_interfaces/environment';
import { ServerService } from 'src/app/_services/server.service';
import { MessageService } from 'primeng/api';

@Component({
    selector: 'app-create-environment',
    templateUrl: './create-environment.component.html',
    styleUrl: './create-environment.component.css'
})
export class CreateEnvironmentComponent {
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
        private clusterService: ClusterService,
        private projectService: ProjectService,
        private serverService: ServerService,
        private environmentService: EnvironmentService,
        private messageService: MessageService,
    ) {
        this.environmentForm = new FormGroup({});
    }

    ngOnInit() {
        this.environment = {
            ID: '',
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

        this.clusterService.getOwnedClusters().subscribe(
            (resp) => { this.clusters = resp; }
        )
        this.projectService.getOwnedProjects().subscribe(
            (resp) => { this.projects = resp; }
        )
    }
    
    get formControls() { return this.environmentForm.controls; }

    onSubmit() {
        this.submitted = true;
        // stop here if form is invalide
        if (this.environmentForm.invalid) {
            return;
        }
        this.environment = { ...this.environment, ...this.environmentForm.value };
        this.environmentService.createEnvironment(this.environment)
            .pipe(first())
            .subscribe({
                next: (resp) => {
                    this.messageService.add({ severity: 'success', summary: "You have successfully created the environment!" });
                    this.router.navigate(['/projects/' + this.environmentForm.controls['ProjectID'].value])
                },
                error: (error: HttpErrorResponse) => {
                    this.messageService.add({ severity: 'error', summary: 'Creation failed', detail: error.error.message });
                    console.error("Environment creation error :", error);
                }
            })
    }

}
