import { Component, ElementRef, ViewChild } from '@angular/core';
import { EnvironmentService } from 'src/app/_services/environment.service';
import { ActivatedRoute, Router } from '@angular/router';
import { Environment } from 'src/app/_interfaces/environment';
import { ClusterService } from 'src/app/_services/cluster.service';
import { Cluster } from 'src/app/_interfaces/cluster';
import { HttpErrorResponse } from '@angular/common/http';
import { Conditions, Microservice } from 'src/app/_interfaces/microservice';
import { User } from 'src/app/_interfaces';
import { MatTableDataSource } from '@angular/material/table';
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import { MessageService } from 'primeng/api';

@Component({
    selector: 'app-environment-details',
    templateUrl: './environment-details.component.html',
    styleUrl: './environment-details.component.css'
})
export class EnvironmentDetailsComponent {

    @ViewChild('closeModal') closeModal!: ElementRef;
    @ViewChild(MatPaginator)
    paginator!: MatPaginator;
    @ViewChild(MatSort)
    sort!: MatSort;
    displayedColumns: string[] = ['Name', 'Strategy', 'Conditions', 'Replicas', 'actions'];
    dataSource: MatTableDataSource<Microservice> = new MatTableDataSource<Microservice>();

    envId: string = '';
    env!: { "environment": Environment };
    cluster!: Cluster;
    user!: { "user": User };
    conditions: Conditions = {
        type: "",
        reason: "",
        message: ""
    }

    clusterTokenExpired: boolean = false;

    constructor(
        private route: ActivatedRoute,
        private environmentService: EnvironmentService,
        private clusterService: ClusterService,
        private messageService: MessageService,
        private router: Router    ) {
    }

    ngOnInit() {
        this.route.paramMap.subscribe(params => {
            const id = params.get('envId');
            if (id !== null) {
                this.envId = id;
                this.loadEnvironmentDetails();
                this.loadMicroservices();
            }
        });
    }

    ClickEnv(row: Microservice) {
        this.router.navigate(['/microservices/' + row.ID]);
    }

    onEnvClick(id: string) {
        this.router.navigate(['environments/' + this.envId + '/microservices/' + id]);
    }

    loadEnvironmentDetails() {
        this.environmentService.getEnvironmentDetails(this.envId)
            .subscribe({
                next: (resp) => {
                    this.env = resp;
                    if (this.env.environment.ClusterID) {
                        this.clusterService.getClusterById(this.env.environment.ClusterID)
                            .subscribe({
                                next: (resp) => {
                                    this.cluster = resp.cluster;
                                    this.clusterTokenExpired = this.clusterService.hasExpired(this.cluster);
                                    console.log(this.clusterTokenExpired);
                                },
                                error: (error: HttpErrorResponse) => {
                                    console.error("Error getting cluster: ", error.error.message || error.error);
                                    this.messageService.add({ severity: 'error', summary: 'Oopss', detail: "Failed to fetch cluster details. Please try again later." });
                                }
                            })
                    }
                },
                error: (error: HttpErrorResponse) => {
                    this.messageService.add({ severity: 'error', summary: 'Oopss', detail: "Failed to fetch environment details. Please try again later." });
                    console.error("Error environment cluster: ", error.error.message);
                }
            });
    }

    confirmUpdate(): void {
        this.environmentService.updateEnvironment(this.env.environment).subscribe(() => {
            this.closeModal.nativeElement.click();
        });
    }

    loadMicroservices() {
        this.environmentService.getMicroservices(this.envId)
            .subscribe({
                next: (resp) => {
                    this.dataSource.data = resp.microservices as Microservice[];
                    this.dataSource.paginator = this.paginator;
                    this.dataSource.sort = this.sort;
                },
                error: (error: HttpErrorResponse) => {
                    this.messageService.add({ severity: 'error', summary: 'Oopss', detail: "Failed to fetch microservices. Please try again later." });
                    console.log(error);
                }
            });

    }
}
