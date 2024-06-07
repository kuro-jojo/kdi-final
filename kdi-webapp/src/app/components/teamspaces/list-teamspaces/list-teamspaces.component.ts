import { Component, ViewChild } from '@angular/core';
import { Teamspace } from 'src/app/_interfaces/teamspace';
import { TeamspaceService } from 'src/app/_services/teamspace.service';
import { HttpErrorResponse } from '@angular/common/http';
import { MatTableDataSource } from '@angular/material/table';
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import { UserService } from 'src/app/_services';
import { ServerService } from 'src/app/_services/server.service';
import { MessageService } from 'primeng/api';


@Component({
    selector: 'app-list-teamspaces',
    templateUrl: './list-teamspaces.component.html',
    styleUrl: './list-teamspaces.component.css'
})
export class ListTeamspacesComponent {


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
        private teamspaceService: TeamspaceService,
        private userService: UserService,
        private serverService: ServerService,
        private messageService: MessageService,
    ) {
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
                            this.messageService.add({ severity: 'info', summary: "Failed to fetch teamspaces. Please try again later." });
                            console.log("Teamspace fetch error: ", error);
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
                                        }
                                    }

                                )
                            }
                        },
                        error: (_error: HttpErrorResponse) => {
                            this.messageService.add({ severity: 'info', summary: "Failed to fetch joined teamspaces. Please try again later." });
                            console.log("Teamspace fetch error: ", _error);
                        }
                    });
            },
        });
    }
}