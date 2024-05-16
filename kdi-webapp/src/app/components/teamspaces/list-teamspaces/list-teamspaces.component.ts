import { Component, ViewChild } from '@angular/core';
import { Teamspace } from 'src/app/_interfaces/teamspace';
import { Router } from '@angular/router';
import { TeamspaceService } from 'src/app/_services/teamspace.service';
import { HttpErrorResponse } from '@angular/common/http';
import { MatTableDataSource } from '@angular/material/table';
import { ToastComponent } from 'src/app/components/toast/toast.component';
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import { UserService } from 'src/app/_services';
import { ServerService } from 'src/app/_services/server.service';


@Component({
    selector: 'app-list-teamspaces',
    templateUrl: './list-teamspaces.component.html',
    styleUrl: './list-teamspaces.component.css'
})
export class ListTeamspacesComponent {
    @ViewChild(ToastComponent) toastComponent!: ToastComponent;

    displayedColumns: string[] = ['Name', 'Description', 'CreatedAt', 'actions'];
    displayedColumnsforJoined: string[] = ['Name', 'Description', 'CreatedAt', 'CreatorID', 'actions'];
    dataSourceforOwned: MatTableDataSource<Teamspace> = new MatTableDataSource<Teamspace>();
    dataSourceforJoined: MatTableDataSource<Teamspace> = new MatTableDataSource<Teamspace>();
    @ViewChild(MatPaginator)
    paginator!: MatPaginator;
    @ViewChild(MatSort)
    sort!: MatSort;

    user!: { 'user': Teamspace };

    constructor(
        private router: Router,
        private teamspaceService: TeamspaceService,
        private userService: UserService,
        private serverService: ServerService,) {
    }

    onTeamClick(id: string) {
        this.router.navigate(['teamspaces' + id]);
    }

    ngOnInit() {
        this.serverService.serverStatus().subscribe({
            next: () => {
                this.teamspaceService.listTeamspacesOwned()
                    .subscribe({
                        next: (resp) => {
                            this.dataSourceforOwned.data = resp.teamspaces as Teamspace[];
                            this.dataSourceforOwned.paginator = this.paginator;
                            this.dataSourceforOwned.sort = this.sort;

                        },
                        error: (error: HttpErrorResponse) => {
                            this.toastComponent.message = "Failed to fetch teamspaces. Please try again later.";
                            this.toastComponent.toastType = 'info';
                            this.triggerToast();
                            console.log(error);
                        },
                        complete: () => {
                            console.log("Teamspaces loaded successfully");
                        }
                    });

                this.teamspaceService.listTeamspacesJoined()
                    .subscribe({
                        next: (resp) => {
                            this.dataSourceforJoined.data = resp.teamspaces as Teamspace[];
                            this.dataSourceforJoined.paginator = this.paginator;
                            this.dataSourceforJoined.sort = this.sort;
                            for (let i = 0; i <= this.dataSourceforJoined.data.length - 1; i++) {
                                this.userService.getUserById(this.dataSourceforJoined.data[i].CreatorID).subscribe(
                                    {
                                        next: (resp) => {
                                            this.user = resp;
                                            this.dataSourceforJoined.data[i].CreatorID = this.user.user.Name;
                                        },
                                        error: (error: HttpErrorResponse) => {
                                            console.error(error.error.message);
                                        },
                                        complete: () => {
                                            console.log("user loaded successfully");
                                        }
                                    }

                                )
                            }
                        },
                        error: (_error: HttpErrorResponse) => {
                            this.toastComponent.message = "Failed to fetch teamspaces. Please try again later.";
                            this.toastComponent.toastType = 'info';
                            this.triggerToast();
                        },
                        complete: () => {
                            console.log("Teamspaces loaded successfully");
                        }
                    });
            },
            error: (_error: HttpErrorResponse) => {
                this.toastComponent.message = "Failed to fetch teamspaces. Please try again later.";
                this.toastComponent.toastType = 'info';
                this.triggerToast();
            }
        });
    }

    triggerToast(): void {
        this.toastComponent.showToast();
    }

}