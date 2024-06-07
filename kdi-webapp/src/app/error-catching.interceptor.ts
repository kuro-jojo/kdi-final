import { Injectable } from '@angular/core';
import { HttpErrorResponse, HttpEvent, HttpHandler, HttpInterceptor, HttpRequest } from '@angular/common/http';
import { Observable, throwError } from 'rxjs';
import { catchError, map } from "rxjs/operators";
import { Router } from '@angular/router';

@Injectable()
export class ErrorCatchingInterceptor implements HttpInterceptor {

    constructor(
        private router: Router
    ) {
    }

    intercept(request: HttpRequest<unknown>, next: HttpHandler): Observable<HttpEvent<unknown>> {
        return next.handle(request)
            .pipe(
                catchError((error: HttpErrorResponse) => {
                    if (!(error.error instanceof ErrorEvent)) {
                        if (error.status === 0) {
                            this.router.navigateByUrl('/500', { state: { redirect: this.router.url } });
                        } if (error.status === 404) {
                            this.router.navigateByUrl('/404');
                        }
                    }
                    return throwError(() => error);
                })
            )
    }
}