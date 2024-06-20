import { HttpErrorResponse } from '@angular/common/http';
import { Component, ElementRef, ViewChild } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import { MatTableDataSource } from '@angular/material/table';
import { ActivatedRoute, Router } from '@angular/router';
import { MessageService } from 'primeng/api';
import { Cluster } from 'src/app/_interfaces/cluster';
import { Environment } from 'src/app/_interfaces/environment';
import { Microservice } from 'src/app/_interfaces/microservice';
import { CacheService } from 'src/app/_services/cache.service';
import { ClusterService } from 'src/app/_services/cluster.service';
import { EnvironmentService } from 'src/app/_services/environment.service';
import { NamespaceService } from 'src/app/_services/namespace.service';
import { ReloadComponent } from 'src/app/component.util';

@Component({
    selector: 'app-microservice-details',
    templateUrl: './microservice-details.component.html',
    styleUrl: './microservice-details.component.css'
})
export class MicroserviceDetailsComponent {

    @ViewChild('overlay') overlay!: ElementRef;
    displayedColumns: string[] = ['Name', 'Image', 'Port', 'actions'];
    displayedLabels: string[] = ['key', 'value',];
    containers: MatTableDataSource<Microservice> = new MatTableDataSource<Microservice>();
    labels: MatTableDataSource<Microservice> = new MatTableDataSource<Microservice>();
    selectors: MatTableDataSource<Microservice> = new MatTableDataSource<Microservice>();
    @ViewChild(MatPaginator)
    paginator!: MatPaginator;
    @ViewChild(MatSort)
    sort!: MatSort;

    microservice!: Microservice
    microserviceId: string = '';
    envId: string = '';
    length: number = 0;
    cluster!: Cluster;
    environment!: { "environment": Environment };
    isModalOpen = false;
    isConfirmModalOpen = false;

    updateForm: FormGroup;
    submitted = false;

    isClusterTokenExpired: boolean = false;


    constructor(
        private clusterService: ClusterService,
        private route: ActivatedRoute,
        private router: Router,
        private environmentService: EnvironmentService,
        private messageService: MessageService,
        private namespaceService: NamespaceService,
        private fb: FormBuilder,
        private cacheService: CacheService,
    ) {
        this.updateForm = new FormGroup({});
    }

    ngOnInit() {
        this.route.paramMap.subscribe(async params => {
            const id = params.get('mId');
            const eId = params.get('envId')
            if (id !== null && eId !== null) {
                this.microserviceId = id;
                this.envId = eId;

                await this.getCluster();
                await this.loadMicroserviceDetails();
                this.length = this.conditionsLength;
                this.generateArray(this.length);

            }
        });

        this.updateForm = this.fb.group({
            image: ['', Validators.required],
            strategy: ['', Validators.required],
            replicas: [1, Validators.required],
            maxUnavailable: [''],
            maxSurge: [''],
            canaryWeight: [''],
            canaryAnalysisInterval: ['']

        });
    }

    get formControls() { return this.updateForm.controls; }

    loadMicroserviceDetails(): Promise<void> {
        return new Promise((resolve, reject) => {
            this.environmentService.getMicroservice(this.envId, this.microserviceId)
                .subscribe({
                    next: (resp) => {
                        this.containers.data = resp.microservice.Containers
                        this.containers.paginator = this.paginator;
                        this.containers.sort = this.sort;
                        this.labels.data = resp.microservice.Labels
                        this.selectors.data = resp.microservice.Selectors
                        this.microservice = resp.microservice;
                        this.namespaceService.getNamespace(this.microservice.namespace).subscribe(
                            {
                                next: (resp) => {
                                    this.microservice.namespace = resp.namespace;
                                    resolve();
                                },
                                error: (error: HttpErrorResponse) => {
                                    console.error("Error loading namespace: ", error.error.message);
                                    reject(error);
                                }
                            })
                    },
                    error: (error: HttpErrorResponse) => {
                        this.messageService.add({ severity: 'info', summary: "Failed to fetch microservice. Please try again later." });
                        console.error("Error loading microservice: ", error.error.message);
                        reject(error);
                    }
                });
        });
    }

    openModal(): void {
        this.isModalOpen = true;
    }

    openConfirmModal(): void {
        this.submitted = true;
        if (this.updateForm.invalid) {
            this.messageService.add({
                severity: 'error',
                summary: 'Form Error',
                detail: 'Please fill in all required fields.'
            });
            return;
        }
        this.isConfirmModalOpen = true;
        this.closeModal();
    }

    closeModal(): void {
        this.isModalOpen = false;
    }

    closeConfirmModal(): void {
        this.isConfirmModalOpen = false;
    }

    isStrategy(strategy: string): boolean {
        return this.updateForm.get('strategy')?.value === strategy;
    }

    onSubmit(): void {
        this.submitted = true;

        if (this.updateForm.valid) {
            this.overlay.nativeElement.style.display = 'block';
            this.closeConfirmModal();
            this.environmentService.updateMicroservice(this.updateForm.value, this.envId, this.microserviceId)
                .subscribe({
                    next: (resp) => {
                        this.messageService.add({ severity: 'success', summary: 'You have successfully updated the microservice!', detail: ' ' });
                        /*this.microservice = resp;
                        this.namespaceService.getNamespace(this.microservice.microservice.namespace).subscribe(
                          {
                              next: (resp) => {
                                  this.microservice.microservice.namespace = resp.namespace;
                              },
                              error: (error: HttpErrorResponse) => {
                                  console.error("Error loading namespace: ", error.error.message);
                              }
                          }
                  
                      )*/
                        console.log('Microservice updated', resp)
                    },
                    error: (error: HttpErrorResponse) => {
                        this.overlay.nativeElement.style.display = 'none';
                        this.messageService.add({ severity: 'info', summary: "Failed to update microservice. Please try again later." });
                        console.error("Error loading microservice: ", error.error.message);
                    },
                    complete: () => {
                        this.overlay.nativeElement.style.display = 'none';
                        setTimeout(() => {
                            this.reloadPage();
                        }, 1000);
                    }
                })

        }
    }

    async getCluster(): Promise<void> {
        return new Promise((resolve, reject) => {
            this.environmentService.getEnvironmentDetails(this.envId).subscribe(
                {
                    next: (resp) => {
                        this.environment = resp;
                        this.clusterService.getClusterById(this.environment.environment.ClusterID).subscribe(
                            {
                                next: (resp) => {
                                    this.cluster = resp.cluster;
                                    this.isClusterTokenExpired = this.clusterService.hasExpired(this.cluster);
                                },
                                error: (error: HttpErrorResponse) => {
                                    console.error("Error loading cluster: ", error.error.message);
                                }
                            }
                        )
                        resolve();
                    },
                    error: (error: HttpErrorResponse) => {
                        console.error("Error loading cluster: ", error.error.message);
                        reject(error);
                    }
                }
            )
        });
    }

    reloadPage() {
        ReloadComponent(true, this.router);
    }

    get conditionsLength(): number {
        if (this.microservice) {
            return this.microservice.Conditions?.length ?? 0;
        }
        return 0

    }

    generateArray(n: number): number[] {
        return Array.from({ length: n }, (_, i) => i);
    }

}
