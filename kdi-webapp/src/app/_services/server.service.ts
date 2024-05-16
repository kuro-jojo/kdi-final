import { Injectable } from '@angular/core';
import { HttpClient, HttpErrorResponse } from '@angular/common/http';

import { User } from '../_interfaces';
import { environment } from 'src/environments/environment';
import { Observable, catchError, map } from 'rxjs';


@Injectable({ providedIn: 'root' })
export class ServerService {

    readonly apiUrl = environment.apiUrl;

    constructor(private http: HttpClient,) {
    }

    serverStatus(): Observable<any> {
        return this.http.get(`${this.apiUrl}/health`);
    }
}   