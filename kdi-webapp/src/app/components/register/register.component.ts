import { Component, ViewChild } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { UserService } from 'src/app/_services';
import { first } from 'rxjs';
import { ToastComponent } from 'src/app/components/toast/toast.component';
import { HttpErrorResponse } from '@angular/common/http';
import { MsalService } from '@azure/msal-angular';

@Component({
    selector: 'app-register',
    templateUrl: './register.component.html',
    styleUrls: ['./register.component.css']
})

export class RegisterComponent {
    registerForm: FormGroup;
    loading = false;
    submitted = false;
    @ViewChild(ToastComponent) toastComponent!: ToastComponent;

    constructor(
        private formBuilder: FormBuilder,
        private router: Router,
        private msalAuthService: MsalService, // for Microsoft login
        private userService: UserService,
    ) {
        // redirect to home if already logged in
        if (this.msalAuthService.instance.getAllAccounts().length > 0 || this.userService.isAuthentificated) {
            // get return url from route parameters or default to '/'
            const { redirect } = window.history.state;
            this.router.navigateByUrl(redirect || '');
        }
        this.registerForm = new FormGroup({});
    }

    get formControls() { return this.registerForm.controls; }

    triggerToast(): void {
        this.toastComponent.showToast();
    }

    ngOnInit() {
        this.registerForm = this.formBuilder.group({
            name: ['', Validators.required],
            email: ['', Validators.email],
            password: ['', [Validators.required, Validators.minLength(6)]]
        });
    }


    onSubmit() {
        this.submitted = true;
        // stop here if form is invalide
        if (this.registerForm.invalid) {
            return;
        }

        this.loading = true;
        this.userService.register(this.registerForm.value)
            .pipe(first())
            .subscribe({
                next: (resp) => {
                    this.toastComponent.message = "You have successfully registered!";
                    this.toastComponent.toastType = 'success';
                    this.triggerToast();

                    // get return url from route parameters or default to '/'
                    const { redirect } = window.history.state;
                    this.router.navigateByUrl(redirect || '');
                    this.loading = false;
                },
                error: (error: HttpErrorResponse) => {
                    this.toastComponent.message = error.error.message;
                    if (error.status == 0) {
                        this.toastComponent.message = "Server is not available";
                    }
                    this.toastComponent.toastType = 'danger';
                    this.triggerToast();

                    this.loading = false;
                }
            })
    }
}