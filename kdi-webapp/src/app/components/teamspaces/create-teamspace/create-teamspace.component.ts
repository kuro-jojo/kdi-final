import { Component, OnInit, ViewChild } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { UserService } from 'src/app/_services';
import { TeamspaceService } from 'src/app/_services/teamspace.service';
import { HttpErrorResponse } from '@angular/common/http';
import { first } from 'rxjs';
import { ToastComponent } from 'src/app/components/toast/toast.component';

@Component({
    selector: 'app-create-teamspace',
    standalone: false,
    templateUrl: './create-teamspace.component.html',
    styleUrl: './create-teamspace.component.css'
})
export class CreateTeamspaceComponent implements OnInit {
    @ViewChild(ToastComponent) toastComponent!: ToastComponent;
    teamspaceForm: FormGroup;
    submitted = false;

    constructor(
        private formBuilder: FormBuilder,
        private router: Router,
        private userService: UserService,
        private teamspaceService: TeamspaceService,) {
        this.teamspaceForm = new FormGroup({});
    }

    ngOnInit(): void {

        this.teamspaceForm = this.formBuilder.group({
            name: ['', Validators.required],
            description: ['', Validators.minLength(6)],
        });

    }

    get formControls() { return this.teamspaceForm.controls; }

    triggerToast(): void {
        this.toastComponent.showToast();
    }

    onSubmit() {
        this.submitted = true;
        // stop here if form is invalide
        if (this.teamspaceForm.invalid) {
            return;
        }
        if (this.userService.isAuthentificated) {

            this.teamspaceService.createTeamspace(this.teamspaceForm.value)
                .pipe(first())
                .subscribe({
                    next: () => {
                        this.toastComponent.message = "You have successfully created a project!";
                        this.toastComponent.toastType = 'success';
                        this.triggerToast();
                        this.router.navigate(['teamspaces'])
                    },
                    error: (error: HttpErrorResponse) => {
                        this.toastComponent.message = error.error.message;
                        this.toastComponent.toastType = 'danger';
                        if (error.status == 0) {
                            this.toastComponent.message = "Server is not available";
                            this.toastComponent.toastType = 'info';
                        }
                        this.triggerToast();
                        console.error("Teamspace creation error :" + error.error.message);
                    }
                })
        }
    }

}
