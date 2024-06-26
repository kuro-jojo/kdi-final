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

@Component({
    selector: 'app-add-openshift-cluster',
    standalone: false,
    templateUrl: './openshift-cluster.component.html',
    styleUrl: './openshift-cluster.component.css'
})
export class AddOpenshiftClusterComponent {
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
            Address: '',
            Port: "",
            Token: '',
        };
    }

    ngOnInit() {
        let getTeamspacePromise = this.getTeamspaces();
        this.cluster.ID = this.route.snapshot.params['id'];
        this.isEditMode = !!this.cluster.ID;

        this.clusterForm = this.formBuilder.group({
            Name: ['', Validators.required],
            Description: [''],
            //TODO : check regex
            Address: ['', [Validators.required, Validators.pattern(/^(https?:\/\/)?[a-z\d.-]+(\.[a-z]{2,6})?(\/[^\s:]*)?$/)]],
            Port: ['', Validators.pattern('^(0|[1-9][0-9]{0,3}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])')],
            Token: ['', Validators.required],
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
                        if (this.cluster.Type != 'openshift') {
                            this.router.navigateByUrl('clusters');
                        }
                    },
                    error: (error) => {
                        this.messageService.add({ severity: 'info', summary: 'Cluster not found', detail: ' ' });
                        this.router.navigateByUrl('clusters');
                    }
                });
        }
    }

    get formControls() { return this.clusterForm.controls; }

    onSubmit() {
        this.submitted = true;
        if (this.clusterForm.value.forTeamspace != 'yes') {
            this.formControls['selectedTeamspaces'].setErrors(null);
        }
        if (this.clusterForm.invalid) {
            return;
        }
        let cluster: Cluster = this.cluster;
        cluster = {
            ...cluster, ...{
                Name: this.clusterForm.value.Name,
                Description: this.clusterForm.value.Description,
                Type: "openshift",
                Address: this.clusterForm.value.Address,
                Port: this.clusterForm.value.Port,
                Token: this.clusterForm.value.Token,
                IsGlobal: this.clusterForm.value.forTeamspace == 'all',
                Teamspaces: this.clusterForm.value.selectedTeamspaces.map((teamspace: Teamspace) => teamspace.ID)
            }
        };
        console.log("cluster : ", cluster);
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

        if (this.formControls['Address'].valid && this.formControls['Port'].valid && this.formControls['Token'].valid) {
            this.testLoading = true;
            this.overlay.nativeElement.style.display = 'block';

            this.clusterService.testConnection(this.clusterForm.value)
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
            this.messageService.add({ severity: 'info', summary: "Please fill address, port and token fields" });
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
