import { Component, ElementRef, ViewChild } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { TeamspaceService } from 'src/app/_services/teamspace.service';
import { Teamspace } from 'src/app/_interfaces/teamspace';
import { HttpErrorResponse } from '@angular/common/http';
import { Project } from 'src/app/_interfaces/project';
import { UserService } from 'src/app/_services';
import { MatTableDataSource } from '@angular/material/table';
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import { ProjectService } from 'src/app/_services/project.service';
import { FormBuilder, FormControl, FormGroup, Validators } from '@angular/forms';
import { first } from 'rxjs';
import { User } from 'src/app/_interfaces';
import { Profile } from 'src/app/_interfaces/profile';
import { ProfileService } from 'src/app/_services/profile.service';
import { ReloadComponent } from 'src/app/component.util';
import { MessageService } from 'primeng/api';



@Component({
    selector: 'app-view-teamspace',
    templateUrl: './view-teamspace.component.html',
    styleUrl: './view-teamspace.component.css'
})
export class ViewTeamspaceComponent {
    displayedColumns: string[] = ['Name', 'Description', 'CreatedAt', 'CreatorID'];
    dataSource: MatTableDataSource<Project> = new MatTableDataSource<Project>();
    @ViewChild(MatPaginator)
    paginator!: MatPaginator;
    @ViewChild(MatSort)
    sort!: MatSort;

    @ViewChild('editMember') editMember!: ElementRef;
    @ViewChild('closeModal') closeModal!: ElementRef;

    teamspace!: Teamspace;
    teamId: string = '';
    projects: Project[] = [];
    projectForm: FormGroup;
    submitted = false;
    user!: { "user": User };
    memberEmail: FormControl = new FormControl('', [Validators.required, Validators.email]);
    memberProfile: FormControl = new FormControl('', Validators.required);
    profiles!: Profile[]
    addMemberFormSubmitted = false;
    editMemberFormSubmitted = false;
    removeMemberFormSubmitted = false;
    loading = false;

    constructor(private route: ActivatedRoute,
        private router: Router,
        private formBuilder: FormBuilder,
        private teamService: TeamspaceService,
        private userService: UserService,
        private projectService: ProjectService,
        private profileService: ProfileService,
        private messageService: MessageService,
    ) {
        this.projectForm = new FormGroup({});
    }

    get formControls() { return this.projectForm.controls; }

    ngOnInit() {
        this.route.paramMap.subscribe(params => {
            const id = params.get('teamId');
            //console.log(id)
            if (id !== null) {
                this.teamId = id;
                this.loadTeamDetails();
                this.loadTeamspaceProjects();
                this.getProfiles();
            }
        });

        this.projectForm = this.formBuilder.group({
            name: ['', Validators.required],
            description: [''],
            teamspace_id: [''],
        });
    }

    loadTeamDetails() {
        this.teamService.getTeamDetails(this.teamId)
            .subscribe({
                next: (resp) => {
                    this.teamspace = resp.teamspace;
                },
                error: (error: HttpErrorResponse) => {
                    this.messageService.add({ severity: 'error', summary: "Failed to fetch teamspace. Please try again later." });
                    console.error("Error loading teamspace: ", error);
                }
            });
    }

    loadTeamspaceProjects() {
        this.teamService.getTeamspaceProjects(this.teamId)
            .subscribe({
                next: (resp) => {
                    this.dataSource.data = resp.projects as Project[];
                    this.dataSource.paginator = this.paginator;
                    this.dataSource.sort = this.sort;

                    for (let i = 0; i <= this.dataSource.data.length - 1; i++) {
                        this.userService.getUserById(this.dataSource.data[i].CreatorID).subscribe(
                            {
                                next: (resp) => {
                                    this.user = resp;
                                    this.dataSource.data[i].CreatorID = this.user.user.Name;
                                },
                                error: (error: HttpErrorResponse) => {
                                    this.messageService.add({ severity: 'error', summary: "Error loading user info", detail: error.error.message || "Failed to load user info. Please try again later." });
                                    console.error("Error loading user info: ", error);
                                }
                            }

                        )
                    }

                },
                error: (error: HttpErrorResponse) => {
                    this.messageService.add({ severity: 'error', summary: "Failed to fetch projects. Please try again later." });
                    console.error("Error loading projects: ", error);
                }
            });

    }

    onSubmit() {
        this.submitted = true;
        // stop here if form is invalide
        if (this.projectForm.invalid) {
            return;
        }

        this.projectForm.value.teamspace_id = this.teamId;
        this.projectService.createProject(this.projectForm.value)
            .pipe(first())
            .subscribe({
                next: (resp) => {
                    this.messageService.add({ severity: 'success', summary: "You have successfully created a project!" });
                    location.reload();
                },
                error: (error: HttpErrorResponse) => {
                    this.messageService.add({ severity: 'error', summary: error.error.message });
                    console.error("Project creation error :", error.error.message);
                }
            })
    }

    getProfiles() {
        this.profileService.getProfiles().subscribe({
            next: (resp) => {
                this.profiles = resp.profiles;
            },
            error: (error: HttpErrorResponse) => {
                this.messageService.add({ severity: 'success', summary: "Failed to fetch profiles. Please try again later." });
                console.log("Error loading profiles: ", error);
            }
        });
    }

    showMemberEditModal(event: any) {
        const element = event.target.parentElement;

        const emailText = this.editMember.nativeElement.querySelector('.modal-body #email');
        const memberId = this.editMember.nativeElement.querySelector('.modal-body #memberId');
        emailText.textContent = element.getAttribute('data-bs-email');
        memberId.textContent = element.getAttribute('data-bs-member-id');
    }

    addMember(e: Event) {
        e.preventDefault();
        this.addMemberFormSubmitted = true;
        if (this.memberEmail.invalid || this.memberProfile.invalid) {
            return;
        }
        this.loading = true;
        this.teamService.addMember(this.teamId, this.memberEmail.value, this.memberProfile.value.ID)
            .subscribe({
                next: () => {
                    this.messageService.add({ severity: 'success', summary: "Member added successfully" });
                    this.memberEmail.reset();
                    this.memberProfile.reset();
                    this.addMemberFormSubmitted = false;
                    this.loadTeamDetails();

                    this.loading = false;
                },
                error: (error: HttpErrorResponse) => {
                    this.messageService.add({ severity: 'error', summary: error.error.message || "Failed to add member" });
                    this.loading = false;
                }
            });
    }

    updateMember(e: Event) {
        e.preventDefault();
        this.editMemberFormSubmitted = true;
        if (this.memberProfile.invalid) {
            return;
        }
        const memberId = this.editMember.nativeElement.querySelector('.modal-body #memberId');

        this.loading = true;
        this.teamService.updateMember(this.teamId, memberId.textContent, this.memberProfile.value.ID)
            .subscribe({
                next: () => {
                    this.messageService.add({ severity: 'success', summary: "Member updated successfully" });
                    this.editMemberFormSubmitted = false;
                    this.closeModal.nativeElement.click();
                    this.loadTeamDetails();
                    this.loading = false;
                },
                error: (error: HttpErrorResponse) => {
                    this.messageService.add({ severity: 'error', summary: error.error.message || "Failed to update member" });
                    this.loading = false;
                }
            });
    }

    removeMember(e: Event) {
        e.preventDefault();
        const emailText = this.editMember.nativeElement.querySelector('.modal-body #email');
        const memberId = this.editMember.nativeElement.querySelector('.modal-body #memberId');

        this.removeMemberFormSubmitted = true;
        if (confirm(`Are you sure you want to remove this member (@${emailText.textContent})?`)) {
            this.loading = true;

            this.teamService.removeMember(this.teamId, memberId.textContent)
                .subscribe({
                    next: () => {
                        this.messageService.add({ severity: 'success', summary: "Member removed successfully" });
                        this.addMemberFormSubmitted = false;
                        this.closeModal.nativeElement.click();
                        this.reloadPage();
                        this.loading = false;
                    },
                    error: (error: HttpErrorResponse) => {
                        this.messageService.add({ severity: 'error', summary: error.error.message || "Failed to remove member" });
                        this.loading = false;
                    }
                });
        }
    }


    reloadPage() {
        ReloadComponent(true, this.router);
    }


}

