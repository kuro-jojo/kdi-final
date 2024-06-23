import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';

import { environment } from 'src/environments/environment';
import { Observable, tap } from 'rxjs';
import { Cluster } from '../_interfaces/cluster';
import { CacheService } from './cache.service';


@Injectable({ providedIn: 'root' })
export class ClusterService {

    readonly apiUrl = environment.apiUrl + '/dashboard/clusters';
    constructor(
        private http: HttpClient,
        private cacheService: CacheService,
    ) { }

    testConnection(cluster: Cluster): Observable<any> {
        return this.http.post<Cluster>(this.apiUrl + '/test', cluster)
    }

    addCluster(cluster: Cluster): Observable<any> {
        return this.http.post<Cluster>(this.apiUrl, cluster).pipe(
            tap(() => {
                this.cacheService.deleteAllRelated(this.apiUrl);
            })
        );
    }

    editCluster(cluster: Cluster): Observable<any> {
        return this.http.patch<Cluster>(this.apiUrl + '/' + cluster.ID, cluster).pipe(
            tap(() => {
                this.cacheService.deleteAllRelated(this.apiUrl + '/' + cluster.ID);
            })
        );
    }

    deleteCluster(id: string): Observable<any> {
        return this.http.delete<any>(this.apiUrl + '/' + id).pipe(
            tap(() => {
                this.cacheService.deleteAllRelated(this.apiUrl);
            })
        );
    }

    getOwnedClusters(): Observable<any> {
        return this.http.get<any>(this.apiUrl + '/owned')
    }

    getClusterById(id: string | undefined, forEdit: boolean = false): Observable<any> {
        if (forEdit) {
            // add query parameter to get all the details of the cluster
            return this.http.get<any>(this.apiUrl + '/' + id + '?token=false')
        }
        return this.http.get<any>(this.apiUrl + '/' + id)
    }

    getClusterNameById(id: string): Observable<any> {
        return this.http.get<any>(this.apiUrl + '/Name/' + id)
    }


    getNamespaces(id: string): Observable<any> {
        return this.http.get<any>(this.apiUrl + '/' + id + '/namespaces')
    }

    hasExpired(cluster: Cluster): boolean {
        return cluster.ExpiryDate !== undefined && new Date(cluster.ExpiryDate) < new Date();
    }
}