import { Component, ViewChild } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { ProjectService } from 'src/app/_services/project.service';
import { Router } from '@angular/router';
import { first } from 'rxjs';
import { ToastComponent } from 'src/app/components/toast/toast.component';
import { HttpErrorResponse } from '@angular/common/http';
import { UserService } from 'src/app/_services';
import { TeamspaceService } from 'src/app/_services/teamspace.service';
import { Teamspace } from 'src/app/_interfaces/teamspace';

@Component({
    selector: 'app-create-project',
    templateUrl: './create-project.component.html',
    styleUrl: './create-project.component.css'
})
export class CreateProjectComponent {
    @ViewChild(ToastComponent) toastComponent!: ToastComponent;
    projectForm: FormGroup;
    submitted = false;
    teamspaces: { teamspaces: Teamspace[], size: number } = { teamspaces: [], size: 0 };

    constructor(
        private formBuilder: FormBuilder,
        private router: Router,
        private userService: UserService,
        private teamService: TeamspaceService,
        private projectService: ProjectService,) {
        this.projectForm = new FormGroup({});
    }

    triggerToast(): void {
        this.toastComponent.showToast();
    }

    chooseTeamsapce($event: Event) {
        console.log($event.target);
        this.projectForm.controls['teamspace'].setValue($event.target ? ['value'] : '');
    }
    ngOnInit() {

        this.projectForm = this.formBuilder.group({
            name: ['', Validators.required],
            description: ['', Validators.minLength(6)],
            teamspace_id: [''],
        });
        this.teamService.listTeamspacesOwned().subscribe(
            (resp) => { this.teamspaces = resp; }
        )
    }
    get formControls() { return this.projectForm.controls; }

    onSubmit() {
        this.submitted = true;
        // stop here if form is invalide
        if (this.projectForm.invalid) {
            return;
        }

        const formData = this.projectForm.value;

        // Si Teamspace n'est pas sélectionné, retirez-le du formulaire
        if (!formData.teamspace) {
            delete formData.teamspace;
        }
        if (this.userService.isAuthentificated) {

            this.projectService.createProject(formData)
                .pipe(first())
                .subscribe({
                    next: (resp) => {
                        this.toastComponent.message = "You have successfully created a project!";
                        this.toastComponent.toastType = 'success';
                        this.triggerToast();
                        this.router.navigate(['/projects'])
                    },
                    error: (error: HttpErrorResponse) => {
                        this.toastComponent.message = error.error.message;
                        this.toastComponent.toastType = 'danger';
                        if (error.status == 0) {
                            this.toastComponent.message = "Server is not available";
                            this.toastComponent.toastType = 'info';
                        }
                        this.triggerToast();

                        console.error("Project creation error :" + error.error.message);
                        console.log('le formulaire', this.projectForm.controls)
                    },
                    complete: () => {
                        console.log("Project created successfully");
                    }
                })
        } else {
            this.toastComponent.message = 'Token invalide';
            this.toastComponent.toastType = 'danger';
        }
    }
}