import { Component } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { UserService } from 'src/app/_services';
import { Router } from '@angular/router';
import { first, timer } from 'rxjs';
import { MsalService } from '@azure/msal-angular';
import { HttpErrorResponse } from '@angular/common/http';
import { MessageService } from 'primeng/api';

@Component({
    selector: 'app-login',
    templateUrl: './login.component.html',
    styleUrls: ['./login.component.css']
})
export class LoginComponent {
    loginForm: FormGroup;
    submitted = false;
    loading = false;
    isIframe = false;

    constructor(
        private formBuilder: FormBuilder,
        private router: Router,
        private userService: UserService,
        private msalAuthService: MsalService, // for Microsoft login
        private messageService: MessageService,
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
                    this.messageService.add({ severity: 'success', summary: 'You have successfully logged in!', detail: ' ' });
                    timer(1000).subscribe(() => {
                        const { redirect } = window.history.state;
                        this.router.navigateByUrl(redirect || '');
                    });
                },
                error: (error: HttpErrorResponse) => {
                    this.messageService.add({ severity: 'error', summary: 'Login failed', detail: error.error.message || "Invalid email or password" });
                    console.error("Login user error :", error.message);
                    this.loading = false;
                }
            })
    }
}

