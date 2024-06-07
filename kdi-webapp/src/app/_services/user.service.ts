import { Injectable } from '@angular/core';
import { HttpClient, HttpErrorResponse } from '@angular/common/http';

import { User } from '../_interfaces';
import { environment } from 'src/environments/environment';
import { Observable, catchError, map } from 'rxjs';


@Injectable({ providedIn: 'root' })
export class UserService {

    readonly apiUrl = environment.apiUrl;
    private userToken: string | null | undefined;
    private readonly tokenKey = 'api.accessToken';

    constructor(private http: HttpClient,) {
        this.userToken = localStorage.getItem(this.tokenKey);
    }

    public get isAuthentificated(): boolean {
        return this.userToken !== null && this.userToken !== undefined;
    }

    public get token(): string | null | undefined {
        return this.userToken;
    }

    public set token(token: string) {
        this.userToken = token;
        localStorage.setItem(this.tokenKey, token);
    }

    getCurrentUser(): Observable<any> {
        return this.http.get<User>(this.apiUrl + `/dashboard/users/current`)
    }

    getUserById(id: string): Observable<any> {
        return this.http.get<User>(this.apiUrl + `/dashboard/users/` + id)
    }

    register(user: User): Observable<any> {
        return this.http.post<User>(this.apiUrl + `/register`, user)
    }

    registerUserWithMsal(): Observable<any> {
        return this.http.post<any>(this.apiUrl + `/register/msal`, {})
    }

    login(user: User): Observable<any> {
        return this.http.post<any>(this.apiUrl + `/login`, user)
            .pipe(
                map(resp => {
                    if (resp && resp.token != "") {
                        this.token = resp.token;
                        return resp;
                    }
                    catchError((error: HttpErrorResponse) => {
                        throw error;
                    });
                })
            );
    }

    logout() {
        if (this.userToken != null) {
            this.userToken = null;
            // localStorage.removeItem(this.tokenKey);
            localStorage.clear();
        }
    }
}