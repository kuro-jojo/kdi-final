import { Component, ViewChild } from '@angular/core';
import { UserService } from 'src/app/_services/user.service';
import { ProjectService } from 'src/app/_services/project.service';
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import { MatTableDataSource } from '@angular/material/table';
import { Router } from '@angular/router';
import { first, timer } from 'rxjs';
import { HttpErrorResponse } from '@angular/common/http';
import { Project } from 'src/app/_interfaces/project';
import { TeamspaceService } from 'src/app/_services/teamspace.service';
import { Teamspace } from 'src/app/_interfaces/teamspace';
import { User } from 'src/app/_interfaces';
import { ReloadComponent } from 'src/app/component.util';
import { MessageService } from 'primeng/api';



@Component({
    selector: 'app-list-projects',
    templateUrl: './list-projects.component.html',
    styleUrl: './list-projects.component.css'
})
export class ListProjectsComponent {
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
        private messageService: MessageService,
    ) {
    }

    viewProject(id: string) {
        this.router.navigate(['projects/' + id]);
    }

    ngOnInit() {
        this.getOwnedProjects();
        this.getJoinedProjects();
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
                                    this.dataSourceforJoined.data[i].Owner = resp.user;
                                },
                                error: (error: HttpErrorResponse) => {
                                    console.error("Error loading user: ", error.error.message, this.dataSourceforJoined.data[i].CreatorID);
                                }
                            }

                        );
                    }
                },
                error: (error: HttpErrorResponse) => {
                    this.messageService.add({ severity: 'error', summary: error.error.message });
                }
            });
    }

    private getOwnedProjects() {
        this.projectService.getOwnedProjects()
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
                    this.messageService.add({ severity: 'info', summary: "Failed to fetch projects. Please try again later." });
                }
            });
    }

    getProjectName(projectId: string) {
        this.projectService.getProjectDetails(projectId).subscribe((resp) => {
            this.project = resp;
            return this.project.Name;

        });
    };

    deleteProject(projectId: string): void {
        if (confirm('Are you sure you want to delete this project?')
            && confirm('This action is irreversible. All data related to this project will be lost.')
            && confirm('Are you sure you want to delete this project?')) {
            this.projectService.deleteProject(projectId).subscribe({
                next: () => {
                    console.log("Project deleted successfully!");
                    this.messageService.add({ severity: 'info', summary: "Project deleted successfully!" });
                    // Rechargement de la liste des projets après la suppression
                    timer(1000).subscribe(() => {
                        this.reloadPage();
                    });
                },
                error: (error: HttpErrorResponse) => {
                    this.messageService.add({ severity: 'error', summary: error.error.message });
                }
            });
        }
    }

    reloadPage() {
        ReloadComponent(true, this.router);
    }
}
