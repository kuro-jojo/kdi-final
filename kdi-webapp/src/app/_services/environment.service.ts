import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Environment } from 'src/app/_interfaces/environment';
import { Observable } from 'rxjs';
import { environment } from 'src/environments/environment';

@Injectable({
    providedIn: 'root'
})
export class EnvironmentService {

    readonly API_URL = environment.apiUrl + '/dashboard/environments';
    constructor(
        private http: HttpClient,
    ) { }

    createEnvironment(environment: Environment): Observable<any> {
        return this.http.post<Environment>(this.API_URL, environment)
    }

    getenvironments(): Observable<any> {
        return this.http.get<any>(this.API_URL + '/ByCluster')
    }

    getlistProjectEnvironments(id: string): Observable<any> {
        return this.http.get<any>(this.API_URL + '/projects/' + id)
    }

    updateEnvironment(environment: Environment): Observable<any> {
        return this.http.patch<Environment>(this.API_URL + '/' + environment.ID, environment)
    }

    deleteEnvironment(id: string): Observable<any> {
        return this.http.delete(this.API_URL + '/' + id);
    }

    getEnvironmentDetails(id: string): Observable<any> {
        return this.http.get<any>(this.API_URL + '/' + id)
    }

    getMicroservices(id: string): Observable<any> {
        return this.http.get<any>(this.API_URL + '/' + id + '/microservices')
    }
}
