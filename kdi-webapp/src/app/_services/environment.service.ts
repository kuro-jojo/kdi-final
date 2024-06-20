import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Environment } from '../_interfaces/environment';
import { Observable, tap } from 'rxjs';
import { environment } from 'src/environments/environment';
import { UpdateForm } from '../_interfaces/updateForm';
import { CacheService } from './cache.service';

@Injectable({
    providedIn: 'root'
})
export class EnvironmentService {

    readonly apiUrl = environment.apiUrl + '/dashboard/environments';
    constructor(
        private http: HttpClient,
        private cacheService: CacheService,
    ) { }

    createEnvironment(environment: Environment): Observable<any> {
        return this.http.post<Environment>(this.apiUrl, environment).pipe(
            tap(() => {
                this.cacheService.deleteAllRelated(this.apiUrl);
            })
        );
    }

    getenvironments(): Observable<any> {
        return this.http.get<any>(this.apiUrl + '/ByCluster')
    }

    getlistProjectEnvironments(id: string): Observable<any> {
        return this.http.get<any>(this.apiUrl + '/projects/' + id)
    }

    updateEnvironment(environment: Environment): Observable<any> {
        return this.http.patch<Environment>(this.apiUrl + '/' + environment.ID, environment).pipe(
            tap(() => {
                this.cacheService.deleteAllRelated(this.apiUrl + '/' + environment.ID);
            })
        );
    }

    deleteEnvironment(id: string): Observable<any> {
        return this.http.delete(this.apiUrl + '/' + id).pipe(
            tap(() => {
                this.cacheService.deleteAllRelated(this.apiUrl);
            })
        );
    }

    getEnvironmentDetails(id: string): Observable<any> {
        return this.http.get<any>(this.apiUrl + '/' + id)
    }

    getMicroservices(id: string): Observable<any> {
        return this.http.get<any>(this.apiUrl + '/' + id + '/microservices')
    }

    getMicroservice(envId: string, mId: string): Observable<any> {
        return this.http.get<any>(this.apiUrl + '/' + envId + '/microservices/' + mId)
    }

    updateMicroservice(form: UpdateForm, envId: string, mId: string): Observable<any> {
        return this.http.patch<any>(this.apiUrl + '/' + envId + '/microservices/' + mId, form).pipe(
            tap(() => {
                this.cacheService.deleteAllRelated(this.apiUrl + '/' + envId + '/microservices/' + mId);
            })
        );
    }
}
