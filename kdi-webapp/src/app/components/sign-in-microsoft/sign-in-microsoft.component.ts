import { Component, ViewChild } from '@angular/core';
import { Router } from '@angular/router';
import { MsalBroadcastService, MsalService } from '@azure/msal-angular';
import { EventMessage, EventType, InteractionStatus } from '@azure/msal-browser';
import { Subject, filter, takeUntil } from 'rxjs';
import { UserService } from '../../_services';
import { environment } from 'src/environments/environment';
import { ServerService } from '../../_services/server.service';
import { ToastComponent } from 'src/app/components/toast/toast.component';

@Component({
    selector: 'app-sign-in-microsoft',
    templateUrl: './sign-in-microsoft.component.html',
    styleUrl: './sign-in-microsoft.component.css'
})
export class SignInMicrosoftComponent {
    @ViewChild(ToastComponent) toastComponent!: ToastComponent;

    loading = false;
    isAvailable = false;

    private readonly _destroying$ = new Subject<void>();
    private readonly scopes = environment.scopes;

    constructor(
        private broadcastService: MsalBroadcastService,
        private msalAuthService: MsalService,
        private userService: UserService,
        private router: Router,
        private serverService: ServerService,
    ) {
    }

    isMicrosoftSignInEnabled() {
        return environment.clientId !== '' && environment.redirectUri !== '' && environment.authority !== '';
    }

    ngOnInit() {

        this.msalAuthService.initialize()
        this.serverService.serverStatus()
            .subscribe({
                next: () => {
                    this.broadcastService.inProgress$
                        .pipe(
                            filter((status: InteractionStatus) => status === InteractionStatus.None),
                            takeUntil(this._destroying$)
                        )
                        .subscribe(() => {
                        })

                    this.broadcastService.msalSubject$
                        .pipe(
                            filter((msg: EventMessage) => msg.eventType === EventType.LOGIN_SUCCESS),
                        )
                        .subscribe(() => {
                            // when the user logs in, acquire a token for the API and store it in the session 
                            this.msalAuthService.acquireTokenSilent({
                                account: this.msalAuthService.instance.getAllAccounts()[0],
                                scopes: this.scopes
                            }).subscribe({
                                next: (tokenResponse) => {
                                    this.userService.token = tokenResponse.accessToken;
                                    // Then try to register the user in the backend if it's not already registered
                                    this.userService.registerUserWithMsal()
                                        .subscribe({
                                            next: () => {
                                                // get return url from route parameters or default to '/'
                                                const { redirect } = window.history.state;
                                                this.router.navigateByUrl(redirect || '');
                                            },
                                            error: (error) => {
                                                console.error("Error while registering user: " + error.message)
                                            }
                                        });
                                },
                                error: (error) => {
                                    console.error("Error while acquiring token: " + error);
                                }
                            });
                        });
                },
                error: (error) => {
                    this.toastComponent.message = "Server is not available. Please try again later"
                    this.toastComponent.toastType = 'info';
                    this.triggerToast();
                }
            })
    }

    signIn() {
        if (!this.isMicrosoftSignInEnabled()) {
            this.toastComponent.message = "Microsoft sign-in is not enabled. Please contact the administrator."
            this.toastComponent.toastType = 'info';
            this.triggerToast();
            return;
        }
        this.serverService.serverStatus()
            .subscribe({
                next: () => {
                    this.loading = true;
                    this.msalAuthService.loginPopup()
                        .subscribe({
                            next: () => { },
                            error: () => {
                                this.loading = false;
                            },
                            complete: () => {
                                this.loading = false;
                            }
                        });
                },
                error: (error) => {
                    this.toastComponent.message = "Server is not available. Please try again later"
                    this.toastComponent.toastType = 'info';
                    this.triggerToast();
                }

            });

    }

    triggerToast(): void {
        this.toastComponent.showToast();
    }

    ngOnDestroy(): void {
        this._destroying$.next(undefined);
        this._destroying$.complete();
    }
}