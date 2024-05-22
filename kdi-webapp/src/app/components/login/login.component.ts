import { Component, ViewChild } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { UserService } from 'src/app/_services';
import { Router } from '@angular/router';
import { first } from 'rxjs';
import { ToastComponent } from 'src/app/components/toast/toast.component';
import { MsalService } from '@azure/msal-angular';
import { HttpErrorResponse } from '@angular/common/http';

@Component({
    selector: 'app-login',
    templateUrl: './login.component.html',
    styleUrls: ['./login.component.css']
})
export class LoginComponent {

    @ViewChild(ToastComponent) toastComponent!: ToastComponent;
    loginForm: FormGroup;
    submitted = false;
    loading = false;
    isIframe = false;

    constructor(
        private formBuilder: FormBuilder,
        private router: Router,
        private userService: UserService,
        private msalAuthService: MsalService // for Microsoft login
    ) {
        this.loginForm = new FormGroup({});
    }

    ngOnInit() {
        // redirect to home if already logged in
        if (this.msalAuthService.instance.getAllAccounts().length > 0 || this.userService.isAuthentificated) {
            // get return url from route parameters or default to '/'
            const { redirect } = window.history.state;
            this.router.navigateByUrl(redirect || '');
        }

        this.loginForm = this.formBuilder.group({
            email: ['', Validators.email],
            password: ['', [Validators.required, Validators.minLength(6)]]
        });

        this.isIframe = window !== window.parent && !window.opener;

    }
    get formControls() { return this.loginForm.controls; }

    triggerToast(): void {
        this.toastComponent.showToast();
    }

    onSubmit() {
        this.submitted = true;
        // stop here if form is invalid
        if (this.loginForm.invalid) {
            return;
        }

        this.loading = true;
        this.userService.login(this.loginForm.value)
            .pipe(first())
            .subscribe({
                next: () => {
                    this.toastComponent.message = "You have successfully logged in!";
                    this.toastComponent.toastType = 'success';
                    this.triggerToast();
                    // get return url from route parameters or default to '/'
                    const { redirect } = window.history.state;
                    this.router.navigateByUrl(redirect || '');
                },
                error: (error: HttpErrorResponse) => {
                    this.toastComponent.message = "Invalid email or password";
                    this.toastComponent.toastType = 'danger';
                    if (error.status == 0) {
                        this.toastComponent.message = "Server is not available. Please try again later"
                        this.toastComponent.toastType = 'info';
                    }
                    this.triggerToast();

                    console.error("Login user error :" + error.message);
                    this.loading = false;
                },
                complete: () => {
                    console.log("Login user completed");
                }
            })
    }
}