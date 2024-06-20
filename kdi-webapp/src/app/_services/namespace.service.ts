import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from 'src/environments/environment';
import { CacheService } from './cache.service';

@Injectable({
    providedIn: 'root'
})
export class NamespaceService {

    readonly apiUrl = environment.apiUrl + '/dashboard/namespaces';
    constructor(
        private http: HttpClient,
        private cacheService: CacheService,
    ) { }

    getNamespace(id: string): Observable<any> {
        return this.http.get<any>(this.apiUrl + '/' + id)
    }
}
