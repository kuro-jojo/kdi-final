import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { UserService } from './_services';
import { MsalService } from '@azure/msal-angular';
import { environment } from 'src/environments/environment';

export const AuthGuard: CanActivateFn = (_route, _state) => {
    const userService = inject(UserService);
    const router = inject(Router);
    const msalAuthService = inject(MsalService);
    const scopes = environment.scopes;
    if (!userService.isAuthentificated && msalAuthService.instance.getAllAccounts().length !== 0) {
        msalAuthService.acquireTokenSilent({
            account: msalAuthService.instance.getAllAccounts()[0],
            scopes: scopes
        }).subscribe({
            next: (tokenResponse) => {
                console.log("Token acquired successfully from AuthGuard");
                userService.token = tokenResponse.accessToken;
            },
            error: (error) => {
                console.error("Error while acquiring token: ", error);
                // TODO: redirect to 500 page
                router.navigateByUrl('/login', { state: { redirect: _state.url } });
                throw new Error("Error while acquiring token in AuthGuard");
            }
        });
    }

    if (!userService.isAuthentificated || (userService.token && tokenExpired(userService.token))) {
        router.navigateByUrl('/login', { state: { redirect: _state.url } });
        return false;
    }
    return true;
};

function tokenExpired(token: string) {
    const expiry = (JSON.parse(atob(token.split('.')[1]))).exp;
    return (Math.floor((new Date).getTime() / 1000)) >= expiry;
}