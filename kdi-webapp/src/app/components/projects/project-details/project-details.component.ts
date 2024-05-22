import { Component, ElementRef, ViewChild } from '@angular/core';
import { ToastComponent } from 'src/app/components/toast/toast.component';
import { MatTableDataSource } from '@angular/material/table';
import { Environment } from 'src/app/_interfaces/environment';
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import { ActivatedRoute, Router } from '@angular/router';
import { EnvironmentService } from 'src/app/_services/environment.service';
import { HttpErrorResponse } from '@angular/common/http';
import { ProjectService } from 'src/app/_services/project.service';
import { Project } from 'src/app/_interfaces/project';
import { ClusterService } from 'src/app/_services/cluster.service';
import { Cluster } from 'src/app/_interfaces/cluster';
import { MatSnackBar } from '@angular/material/snack-bar';
import { ReloadComponent } from 'src/app/component.util';
import { TeamspaceService } from 'src/app/_services/teamspace.service';
import { Teamspace } from 'src/app/_interfaces/teamspace';

@Component({
    selector: 'app-project-details',
    templateUrl: './project-details.component.html',
    styleUrl: './project-details.component.css'
})
export class ProjectDetailsComponent {

    @ViewChild(ToastComponent) toastComponent!: ToastComponent;
    @ViewChild('closeModal') closeModal!: ElementRef;
    displayedColumns: string[] = ['Name', 'Description', 'ClusterID', 'actions'];
    dataSource: MatTableDataSource<Environment> = new MatTableDataSource<Environment>();
    @ViewChild(MatPaginator)
    paginator!: MatPaginator;
    @ViewChild(MatSort)
    sort!: MatSort;

    project: { project: Project } = {
        project: {
            ID: '',
            Name: '',
            Description: '',
            CreatedAt: new Date,
            CreatorID: '',
            TeamspaceID: ''
        }
    };
    updatedProject!: Project;
    projectToUpdate!: Project;
    envId: string = '';
    projectId: string = '';
    cluster!: { "cluster": Cluster };
    teamspace!: { "teamspace": Teamspace };
    //environments: Environment[] = [];

    constructor(private route: ActivatedRoute,
        private router: Router,
        private environmentService: EnvironmentService,
        private teamspaceService: TeamspaceService,
        private projectService: ProjectService,
        private clusterService: ClusterService,
        private snackBar: MatSnackBar
    ) {
    }

    ngOnInit() {
        this.route.paramMap.subscribe(params => {
            const id = params.get('projectId');
            if (id !== null) {
                this.projectId = id;
                this.loadProjectDetails();
                this.loadEnvironments();
            }
        });
    }


    loadProjectDetails() {
        this.projectService.getProjectDetails(this.projectId)
            .subscribe({
                next: (resp) => {
                    this.project = resp;
                    if (this.project.project.TeamspaceID) {
                        this.teamspaceService.getTeamspaceById(this.project.project.TeamspaceID).subscribe({
                            next: (resp) => {
                                this.teamspace = resp;
                                this.project.project.TeamspaceID = this.teamspace.teamspace.Name;
                            },
                            error: (error: HttpErrorResponse) => {
                                console.error("Error getting teamspace: ", error.error.message || error.error);
                            }
                        }
                        )
                    }
                },
                error: (error: HttpErrorResponse) => {
                    console.error("Error loading project: ", error.error.message);
                }
            });
    }

    loadEnvironments() {
        this.environmentService.getlistProjectEnvironments(this.projectId)
            .subscribe({
                next: (resp) => {
                    this.dataSource.data = resp.environments as Environment[];
                    this.dataSource.paginator = this.paginator;
                    this.dataSource.sort = this.sort;
                    for (let i = 0; i <= this.dataSource.data.length - 1; i++) {
                        this.clusterService.getClusterById(this.dataSource.data[i].ClusterID).subscribe(
                            {
                                next: (resp) => {
                                    this.cluster = resp;
                                    this.dataSource.data[i].ClusterID = this.cluster.cluster.Name;
                                },
                                error: (error: HttpErrorResponse) => {
                                    console.error("Error loading cluster: ", error.error.message);
                                }
                            }

                        )
                    }

                },
                error: (error: HttpErrorResponse) => {
                    this.toastComponent.message = "Failed to fetch environments. Please try again later.";
                    this.toastComponent.toastType = 'info';
                    this.triggerToast();
                    console.log(error);
                }
            });

    }

    deleteProject(projectId: string): void {
        if (confirm('Are you sure you want to delete this project?')) {
            this.projectService.deleteProject(projectId).subscribe(() => {
                this.snackBar.open('Project deleted successfully', 'Close', {
                    duration: 3000,
                    verticalPosition: 'top',
                    horizontalPosition: 'end'
                });
                // Rechargez la liste des projets après la suppression
                this.reloadPage();
            });
        }
    }
    triggerToast(): void {
        this.toastComponent.showToast();
    }

    confirmUpdate(): void {
        this.projectService.updateProject(this.project.project).subscribe(() => {
            this.closeModal.nativeElement.click();
        });
    }

    reloadPage() {
        ReloadComponent(true, this.router);
    }
}