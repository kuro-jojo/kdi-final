import { Location } from '@angular/common';
import { Component, ViewChild } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { first, forkJoin } from 'rxjs';
import { Cluster } from 'src/app/_interfaces/cluster';
import { Teamspace } from 'src/app/_interfaces/teamspace';
import { ClusterService } from 'src/app/_services/cluster.service';
import { TeamspaceService } from 'src/app/_services/teamspace.service';
import { ToastComponent } from 'src/app/components/toast/toast.component';

@Component({
    selector: 'app-add-local-cluster',
    standalone: false,
    templateUrl: './local-cluster.component.html',
    styleUrl: './local-cluster.component.css'
})
export class AddLocalClusterComponent {
    isEditMode: boolean = false;
    @ViewChild(ToastComponent) toastComponent!: ToastComponent;
    clusterForm: FormGroup;
    submitted: boolean = false;
    addLoading: boolean = false;
    editLoading: boolean = false;
    revokeLoading: boolean = false;
    cluster: Cluster;
    teamspaces!: Teamspace[];

    constructor(
        private router: Router,
        private route: ActivatedRoute,
        private formBuilder: FormBuilder,
        private clusterService: ClusterService,
        private location: Location,
        private teamspaceService: TeamspaceService,
    ) {
        this.clusterForm = new FormGroup({});
        this.cluster = {
            ID: '',
            Name: '',
            Description: '',
            IpAddress: '',
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
            IpAddress: ['', [Validators.required, Validators.pattern('^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$')]],
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

                        if ((this.cluster.Teamspaces?.length ?? 0) == this.teamspaces.length) {
                            this.formControls['forTeamspace'].setValue('all');
                            this.formControls['selectedTeamspaces'].setValue(this.teamspaces);
                        } else if ((this.cluster.Teamspaces?.length ?? 0) > 0) {
                            this.formControls['forTeamspace'].setValue('yes');
                            this.formControls['selectedTeamspaces'].setValue(
                                this.teamspaces.filter((teamspace: Teamspace) => this.cluster.Teamspaces?.includes(teamspace.ID))
                            );
                        } else {
                            this.formControls['forTeamspace'].setValue('no');
                        }
                    },
                    error: (error) => {
                        this.toastComponent.message = "Cluster not found";
                        this.toastComponent.toastType = 'info';
                        this.triggerToast();
                        this.router.navigateByUrl('clusters');
                    }
                });
        }
    }

    get formControls() { return this.clusterForm.controls; }

    triggerToast(): void {
        this.toastComponent.showToast();
    }

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
                IpAddress: this.clusterForm.value.IpAddress,
                Port: this.clusterForm.value.Port,
                Token: this.clusterForm.value.Token,
                IsGlobal: this.clusterForm.value.forTeamspace == 'all',
                Teamspaces: this.clusterForm.value.selectedTeamspaces.map((teamspace: Teamspace) => teamspace.ID)
            }
        };
        if (this.isEditMode) {
            this.editLoading;
            this.clusterService.editCluster(cluster)
                // .pipe(first())
                .subscribe({
                    next: (resp: any) => {
                        this.toastComponent.message = resp.message || "Cluster edited successfully";
                        this.toastComponent.toastType = 'success';
                        this.triggerToast();
                        this.router.navigateByUrl('clusters')
                    },
                    error: (error) => {
                        this.toastComponent.message = error.error.message || "Error editing cluster";
                        this.toastComponent.toastType = 'danger';
                        if (error.status == 0) {
                            this.toastComponent.message = "Server is not available";
                            this.toastComponent.toastType = 'info';
                        }
                        this.triggerToast();
                        this.addLoading = false;
                        console.error("Error adding cluster :" + error.error.message);
                    }
                })
        } else {
            this.addLoading = true;
            this.clusterService.addCluster(cluster)
                .pipe(first())
                .subscribe({
                    next: (resp) => {
                        this.toastComponent.message = resp.message || "Cluster added successfully";
                        this.toastComponent.toastType = 'success';
                        this.triggerToast();
                        this.router.navigateByUrl('clusters')
                    },
                    error: (error) => {
                        this.toastComponent.message = error.error.message || "Error adding cluster";
                        this.toastComponent.toastType = 'danger';
                        if (error.status == 0) {
                            this.toastComponent.message = "Server is not available";
                            this.toastComponent.toastType = 'info';
                        }
                        this.triggerToast();
                        this.editLoading = false;
                        console.error("Error adding cluster :" + error.error.message);
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
                        this.toastComponent.message = "Cluster deleted successfully";
                        this.toastComponent.toastType = 'success';
                        this.triggerToast();
                        this.router.navigateByUrl('clusters')
                    },
                    error: (error) => {
                        this.toastComponent.message = error.error.message || "Error deleting cluster";
                        this.toastComponent.toastType = 'danger';
                        if (error.status == 0) {
                            this.toastComponent.message = "Server is not available";
                            this.toastComponent.toastType = 'info';
                        }
                        this.triggerToast();
                        this.revokeLoading = false;
                        console.error("Error deleting cluster :" + error.error.message);
                    }
                })
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