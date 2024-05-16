
import { Injectable } from '@angular/core';
import { HttpClient, HttpErrorResponse } from '@angular/common/http';

import { environment } from 'src/environments/environment';
import { Observable, catchError, map } from 'rxjs';
import { Cluster } from '../_interfaces/cluster';


@Injectable({ providedIn: 'root' })
export class ClusterService {

    readonly apiUrl = environment.apiUrl + '/dashboard/clusters';
    constructor(
        private http: HttpClient,
    ) { }

    addCluster(cluster: Cluster): Observable<any> {
        return this.http.post<Cluster>(this.apiUrl, cluster)
    }

    editCluster(cluster: Cluster): Observable<any> {
        return this.http.patch<Cluster>(this.apiUrl + '/' + cluster.ID, cluster)
    }

    deleteCluster(id: string): Observable<any> {
        return this.http.delete<any>(this.apiUrl + '/' + id)
    }

    getClusters(): Observable<any> {
        return this.http.get<any>(this.apiUrl + '/owned')
    }

    getClusterById(id: string | undefined) {
        return this.http.get<any>(this.apiUrl + '/' + id)
    }
}