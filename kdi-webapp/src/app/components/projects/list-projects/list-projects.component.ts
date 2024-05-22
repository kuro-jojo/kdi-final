import { Component, ViewChild } from '@angular/core';
import { ToastComponent } from 'src/app/components/toast/toast.component';
import { UserService } from 'src/app/_services/user.service';
import { ProjectService } from 'src/app/_services/project.service';
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import { MatTableDataSource } from '@angular/material/table';
import { Router } from '@angular/router';
import { first } from 'rxjs';
import { HttpErrorResponse } from '@angular/common/http';
import { Project } from 'src/app/_interfaces/project';
import { TeamspaceService } from 'src/app/_services/teamspace.service';
import { Teamspace } from 'src/app/_interfaces/teamspace';
import { User } from 'src/app/_interfaces';
import { ReloadComponent } from 'src/app/component.util';
import { ServerService } from 'src/app/_services/server.service';



@Component({
    selector: 'app-list-projects',
    templateUrl: './list-projects.component.html',
    styleUrl: './list-projects.component.css'
})
export class ListProjectsComponent {
    @ViewChild(ToastComponent) toastComponent!: ToastComponent;
    displayedColumns: string[] = ['Name', 'Description', 'CreatedAt', 'Teamspace', 'actions'];
    displayedColumnsforJoined: string[] = ['Name', 'Description', 'CreatedAt', 'Teamspace', 'CreatorID', 'actions'];
    dataSource: MatTableDataSource<Project> = new MatTableDataSource<Project>();
    dataSourceforJoined: MatTableDataSource<Project> = new MatTableDataSource<Project>();
    @ViewChild(MatPaginator)
    paginator!: MatPaginator;
    @ViewChild(MatSort)
    sort!: MatSort;

    teamspaces: Record<string, Teamspace> = {};
    teamspacejoined!: Teamspace
    user!: { "user": User };
    project!: Project;
    projectToDelete!: Project;
    formBuilder: any;


    constructor(
        private router: Router,
        private projectService: ProjectService,
        private teamspaceService: TeamspaceService,
        private userService: UserService,
        private serverService: ServerService,
    ) {
    }

    onProjectClick(id: string) {
        this.router.navigate(['projects/' + id]);
    }

    ngOnInit() {
        this.serverService.serverStatus()
            .subscribe({
                next: () => {
                    this.getOwnedProjects();

                    this.getJoinedProjects();
                },
                error: (error: HttpErrorResponse) => {
                    this.toastComponent.message = "Server is not available. Please try again later"
                    this.toastComponent.toastType = 'info';
                    this.triggerToast();
                    //console.log(error);
                }
            });

    }

    private getJoinedProjects() {
        this.projectService.listProjectsOfJoinedTeamspaces()
            .pipe(first())
            .subscribe({
                next: (resp) => {
                    this.dataSourceforJoined.data = resp.projects as Project[];
                    this.dataSourceforJoined.paginator = this.paginator;
                    this.dataSourceforJoined.sort = this.sort;

                    for (let i = 0; i <= this.dataSourceforJoined.data.length - 1; i++) {
                        this.teamspaceService.getTeamspaceById(this.dataSourceforJoined.data[i].TeamspaceID).subscribe(
                            {
                                next: (resp) => {
                                    this.teamspaces[this.dataSourceforJoined.data[i].ID] = resp.teamspace;
                                },
                                error: (error: HttpErrorResponse) => {
                                    console.error("Error loading teamspace: ", error.error.message);
                                }
                            }

                        );
                    }

                    for (let i = 0; i <= this.dataSourceforJoined.data.length - 1; i++) {
                        this.userService.getUserById(this.dataSourceforJoined.data[i].CreatorID).subscribe(
                            {
                                next: (resp) => {
                                    this.user = resp;
                                    this.dataSourceforJoined.data[i].CreatorID = resp.user;
                                },
                                error: (error: HttpErrorResponse) => {
                                    console.error("Error loading user: ", error.error.message);
                                }
                            }

                        );
                    }
                },
                error: (error: HttpErrorResponse) => {
                    this.toastComponent.message = error.error.message;
                    this.toastComponent.toastType = 'danger';
                    this.triggerToast();

                }
            });
    }

    private getOwnedProjects() {
        this.projectService.listProjects()
            .subscribe({
                next: (resp) => {
                    this.dataSource.data = resp.projects as Project[];
                    this.dataSource.paginator = this.paginator;
                    this.dataSource.sort = this.sort;
                    for (let i = 0; i <= this.dataSource.data.length - 1; i++) {
                        if (this.dataSource.data[i].TeamspaceID) {
                            this.teamspaceService.getTeamspaceById(this.dataSource.data[i].TeamspaceID).subscribe(
                                {
                                    next: (resp) => {
                                        this.teamspaces[this.dataSource.data[i].ID] = resp.teamspace;
                                    },
                                    error: (error: HttpErrorResponse) => {
                                        console.error("Error loading teamspace: ", error.error.message);
                                    }
                                }

                            );
                        }
                    }
                },
                error: (error: HttpErrorResponse) => {
                    this.toastComponent.message = "Failed to fetch projects. Please try again later.";
                    this.toastComponent.toastType = 'info';
                    this.triggerToast();
                }
            });
    }

    triggerToast(): void {
        this.toastComponent.showToast();
    }

    getProjectName(projectId: string) {
        this.projectService.getProjectDetails(projectId).subscribe((resp) => {
            this.project = resp;
            return this.project.Name;

        });
    };

    deleteProject(projectId: string): void {
        if (confirm('Are you sure you want to delete this project?')) {
            this.projectService.deleteProject(projectId).subscribe(() => {
                this.toastComponent.message = "Project deleted successfully!";
                this.toastComponent.toastType = 'success';
                this.triggerToast();
                // Rechargement de la liste des projets après la suppression
                this.reloadPage();
            });
        }
    }

    /*deleteProjectFromTeamspace(teamId: string, projectId: string): void{
      if (confirm('Are you sure you want to delete this project?')) {
        this.teamspaceService.deleteProjectInTeamspace(teamId, projectId).subscribe(() => {
          this.toastComponent.message = "Project deleted successfully!";
          this.toastComponent.toastType = 'success';
          this.triggerToast();
            // Rechargement de la liste des projets après la suppression
            this.router.navigateByUrl('/refresh', { skipLocationChange: true }).then(() => {
        this.router.navigate(['/projects']);
    });
        });
    }
    }*/
    confirmDeleteProject(project: any) {
        this.projectToDelete = project;

    }

    reloadPage() {
        ReloadComponent(true, this.router);
    }
}