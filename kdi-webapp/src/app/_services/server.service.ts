import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';

import { environment } from 'src/environments/environment';
import { Observable } from 'rxjs';


@Injectable({ providedIn: 'root' })
export class ServerService {

    readonly apiUrl = environment.apiUrl;

    constructor(private http: HttpClient,) {
    }

    serverStatus(): Observable<any> {
        return this.http.get(`${this.apiUrl}/health`);
    }
}   