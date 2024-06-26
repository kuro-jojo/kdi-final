import { Location } from '@angular/common';
import { Component, ElementRef, ViewChild } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { MessageService } from 'primeng/api';
import { first, forkJoin, timer } from 'rxjs';
import { Cluster } from 'src/app/_interfaces/cluster';
import { Teamspace } from 'src/app/_interfaces/teamspace';
import { ClusterService } from 'src/app/_services/cluster.service';
import { TeamspaceService } from 'src/app/_services/teamspace.service';
import * as regionsJson from 'src/assets/data/regions.json';

@Component({
    selector: 'app-add-eks-cluster',
    standalone: false,
    templateUrl: './eks-cluster.component.html',
    styleUrl: './eks-cluster.component.css'
})
export class AddEksClusterComponent {
    @ViewChild('overlay') overlay!: ElementRef;
    isEditMode: boolean = false;
    clusterForm: FormGroup;
    submitted: boolean = false;
    addLoading: boolean = false;
    editLoading: boolean = false;
    revokeLoading: boolean = false;
    testLoading: boolean = false;
    cluster: Cluster;
    teamspaces!: Teamspace[];
    regions!: any[]
    isAccessKeyVisible: boolean = false;
    isSecretKeyVisible: boolean = false;

    constructor(
        private router: Router,
        private route: ActivatedRoute,
        private formBuilder: FormBuilder,
        private clusterService: ClusterService,
        private location: Location,
        private teamspaceService: TeamspaceService,
        private messageService: MessageService,
    ) {
        this.clusterForm = new FormGroup({});
        this.cluster = {
            ID: '',
            Name: '',
            Description: '',
            Type: 'eks',
            Region: '',
            AccessKeyID: '',
            SecretKey: '',
        };
    }

    ngOnInit() {
        let getTeamspacePromise = this.getTeamspaces();

        this.regions = Object.keys(regionsJson);
        this.cluster.ID = this.route.snapshot.params['id'];
        this.isEditMode = !!this.cluster.ID;

        this.clusterForm = this.formBuilder.group({
            Name: ['', Validators.required],
            Description: [''],
            //TODO : check regex
            AccessKeyID: ['',
                [
                    Validators.required,
                    Validators.minLength(20),
                    Validators.maxLength(20),
                    Validators.pattern(/^(AKIA|ASIA)[A-Z0-9]{16}$/)
                ]
            ],
            SecretKey: [
                '',
                [
                    Validators.required,
                    Validators.minLength(40),
                    Validators.maxLength(40),
                    Validators.pattern(/^[A-Za-z0-9\/+=]{40}$/)
                ]
            ],
            selectedRegion: [undefined, Validators.required],
            selectedTeamspaces: [[], Validators.required],
            forTeamspace: ['no', Validators.required]
        });
        if (this.isEditMode) {

            this.clusterService.getClusterById(this.cluster.ID, true)
                .pipe(first())
                .subscribe({
                    next: async (resp: any) => {
                        await getTeamspacePromise;
                        this.cluster = resp.cluster;
                        this.clusterForm.patchValue(this.cluster);

                        if (this.cluster.Teamspaces?.length == 0 || this.cluster.Teamspaces == null) {
                            this.formControls['forTeamspace'].setValue('no');
                        } else if ((this.cluster.Teamspaces?.length ?? 0) == this.teamspaces.length) {
                            this.formControls['forTeamspace'].setValue('all');
                            this.formControls['selectedTeamspaces'].setValue(this.teamspaces);
                        } else if ((this.cluster.Teamspaces?.length ?? 0) > 0) {
                            this.formControls['forTeamspace'].setValue('yes');
                            this.formControls['selectedTeamspaces'].setValue(
                                this.teamspaces.filter((teamspace: Teamspace) => this.cluster.Teamspaces?.includes(teamspace.ID))
                            );
                        }
                        if (this.cluster.Type != 'eks') {
                            this.router.navigateByUrl('clusters');
                        }
                        this.formControls['selectedRegion'].setValue(this.cluster.Region);
                    },
                    error: (error) => {
                        this.messageService.add({ severity: 'info', summary: 'Cluster not found', detail: ' ' });
                        this.router.navigateByUrl('clusters');
                    }
                });
        }
    }

    get formControls() { return this.clusterForm.controls; }

    toggleAccessKeyVisibility() {
        this.isAccessKeyVisible = !this.isAccessKeyVisible;
    }

    toggleSecretKeyVisibility() {
        this.isSecretKeyVisible = !this.isSecretKeyVisible;
    }
    onSubmit() {
        this.submitted = true;
        if (this.clusterForm.value.forTeamspace != 'yes') {
            this.formControls['selectedTeamspaces'].setErrors(null);
        }
        if (this.clusterForm.invalid) {
            return;
        }
        let cluster: Cluster = this.getClusterToSubmit();

        if (this.isEditMode) {
            this.editLoading;
            this.clusterService.editCluster(cluster)
                // .pipe(first())
                .subscribe({
                    next: (resp: any) => {
                        this.messageService.add({ severity: 'info', summary: resp.message || "Cluster edited successfully", detail: ' ' });
                        timer(1000).subscribe(() => {
                            this.router.navigateByUrl('clusters')
                        });
                    },
                    error: (error) => {
                        this.messageService.add({ severity: 'error', summary: "Edit failed", detail: error.error.message || "Error editing cluster" });
                        this.editLoading = false;
                        console.error("Error adding cluster :", error.error.message);
                    }
                })
        } else {
            this.addLoading = true;
            this.clusterService.addCluster(cluster)
                .pipe(first())
                .subscribe({
                    next: (resp) => {
                        this.messageService.add({ severity: 'success', summary: resp.message || "Cluster added successfully", detail: ' ' });
                        timer(1000).subscribe(() => {
                            this.router.navigateByUrl('clusters')
                        });
                    },
                    error: (error) => {
                        this.messageService.add({ severity: 'error', summary: "Creation failed", detail: error.error.message || "Error adding cluster" });
                        this.addLoading = false;
                        console.error("Error adding cluster :", error.error.message);
                    }
                })
        }
    }

    private getClusterToSubmit() {
        let cluster: Cluster = this.cluster;
        cluster = {
            ...cluster, ...{
                Name: this.clusterForm.value.Name,
                Description: this.clusterForm.value.Description,
                Type: "eks",
                AccessKeyID: this.clusterForm.value.AccessKeyID,
                SecretKey: this.clusterForm.value.SecretKey,
                Region: this.clusterForm.value.selectedRegion,
                IsGlobal: this.clusterForm.value.forTeamspace == 'all',
                Teamspaces: this.clusterForm.value.selectedTeamspaces.map((teamspace: Teamspace) => teamspace.ID),
                Token: ''
            }
        };
        return cluster;
    }

    revokeCluster() {
        console.log("Revoking cluster");
        this.revokeLoading = true;
        if (this.isEditMode && confirm("Are you sure you want to delete this cluster?") && this.cluster.ID) {
            this.clusterService.deleteCluster(this.cluster.ID)
                .pipe(first())
                .subscribe({
                    next: (resp: any) => {
                        this.messageService.add({ severity: 'info', summary: "Cluster deleted successfully", detail: "" });
                        timer(1000).subscribe(() => {
                            this.router.navigateByUrl('clusters')
                        });
                    },
                    error: (error) => {
                        this.messageService.add({ severity: 'error', summary: "Deletion failed", detail: error.error.message || "Error deleting cluster" });
                        this.revokeLoading = false;
                        console.error("Error deleting cluster :", error.error.message);
                    }
                })
        }
    }

    testConnection() {
        console.log("clusterForm : ", this.clusterForm.value);
        if (this.formControls['AccessKeyID'].valid && this.formControls['SecretKey'].valid && this.formControls['selectedRegion'].valid) {
            this.testLoading = true;
            this.overlay.nativeElement.style.display = 'block';
            let cluster: Cluster = this.getClusterToSubmit();

            this.clusterService.testConnection(cluster)
                .pipe(first())
                .subscribe({
                    next: (resp) => {
                        this.overlay.nativeElement.style.display = 'none';
                        this.messageService.add({ severity: 'success', summary: "Connection successful", detail: ' ' });
                        this.testLoading = false;
                    },
                    error: (error) => {
                        this.overlay.nativeElement.style.display = 'none';
                        this.messageService.add({ severity: 'error', summary: "Connection failed", detail: error.error.message || "Error testing connection" });
                        this.testLoading = false;
                        console.error("Error testing connection :", error.error.message);
                    }
                })
        } else {
            this.messageService.add({ severity: 'info', summary: "Please fill cluster name, access key, secret key and region", detail: ' ' });
        }
    }

    cancel() {
        this.location.back();
    }

    getTeamspaces(): Promise<void> {
        let teamspacesOwned$ = this.teamspaceService.listTeamspacesOwned();
        let teamspacesJoined$ = this.teamspaceService.listTeamspacesJoined();

        return new Promise((resolve, reject) => {
            forkJoin([teamspacesOwned$, teamspacesJoined$]).subscribe({
                next: (results) => {
                    let teamspacesOwned = results[0];
                    let teamspacesJoined = results[1];
                    let teamspaces = [...teamspacesOwned.teamspaces || [], ...teamspacesJoined.teamspaces || []];

                    this.teamspaces = teamspaces;
                    resolve();
                }, error: (error) => {
                    console.error("Error loading teamspaces: ", error.error.message || error.message);
                    reject(error);
                }
            });
        });
    }

    onTeamspaceChange() {
        if (this.clusterForm.value.forTeamspace == 'yes') {
            this.formControls['selectedTeamspaces'].setValidators([Validators.required]);
        } else {
            this.formControls['selectedTeamspaces'].setValue([]);
            this.formControls['selectedTeamspaces'].setValidators(null);
        }
        this.formControls['selectedTeamspaces'].updateValueAndValidity();
    }
}
