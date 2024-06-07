import { Injectable } from '@angular/core';
import { HttpInterceptor, HttpRequest, HttpHandler, HttpEvent } from '@angular/common/http';
import { Observable } from 'rxjs';
import { MsalInterceptor, MsalService } from '@azure/msal-angular';
import { UserService } from './_services';
import { Router } from '@angular/router';
import { environment } from 'src/environments/environment';

@Injectable()
export class AuthInterceptor implements HttpInterceptor {
    private account: any;
    private readonly scopes = environment.scopes;

    constructor(
        private msalAuthService: MsalService, // for Microsoft login
        private userService: UserService,
        private router: Router
    ) {
    }

    intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
        this.account = this.msalAuthService.instance.getAllAccounts()[0];
        if (this.account || req.url.includes('msal')) {
            this.msalAuthService.acquireTokenSilent({
                account: this.msalAuthService.instance.getAllAccounts()[0],
                scopes: this.scopes
            }).subscribe({
                next: (tokenResponse) => {
                    this.userService.token = tokenResponse.accessToken;
                },
                error: (error) => {
                    console.error("Error while acquiring token: ", error);
                }
            });

            req = req.clone({
                setHeaders: {
                    "Auth-method": "msal"
                }
            });
        }
        if (this.userService.isAuthentificated) {
            req = req.clone({
                setHeaders: {
                    Authorization: `Bearer ${this.userService.token}`
                }
            });
            return next.handle(req);
        }
        if (!this.router.url.includes('login') && !this.router.url.includes('register')) {
            this.router.navigateByUrl('/login');
        }
        return next.handle(req);
    }
}
