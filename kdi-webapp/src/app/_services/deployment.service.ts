import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';

import { environment } from 'src/environments/environment';
import { Observable } from 'rxjs';
import { CacheService } from './cache.service';


@Injectable({ providedIn: 'root' })
export class DeploymentService {

    readonly apiUrl = environment.apiUrl + '/dashboard/environments';
    readonly microservicesWithYaml = '/microservices/with-yaml';
    constructor(
        private http: HttpClient,
        private cacheService: CacheService, // TODO
    ) { }

    addDeploymentWithYaml(environmentID: string, files: File[]): Observable<any> {
        const formData = new FormData();
        files.forEach(file => {
            formData.append('files', file);
        });
        return this.http.post<any>(this.apiUrl + '/' + environmentID + this.microservicesWithYaml, formData)
    }
}
