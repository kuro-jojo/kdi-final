import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, tap } from 'rxjs';
import { environment } from 'src/environments/environment';
import { Teamspace } from '../_interfaces/teamspace';
import { CacheService } from './cache.service';

@Injectable({
    providedIn: 'root'
})
export class TeamspaceService {

    readonly apiUrl = environment.apiUrl + '/dashboard/teamspaces';

    constructor(
        private http: HttpClient,
        private cacheService: CacheService,
    ) { }


    createTeamspace(teamspace: Teamspace): Observable<any> {
        return this.http.post<any>(this.apiUrl, teamspace).pipe(
            tap(() => {
                this.cacheService.deleteAllRelated(this.apiUrl);
            })
        );
    }

    listTeamspacesOwned(): Observable<any> {
        return this.http.get<any>(this.apiUrl + '/owned')
    }

    listTeamspacesJoined(): Observable<any> {
        return this.http.get<any>(this.apiUrl + '/joined')
    }

    getTeamDetails(id: string): Observable<any> {
        return this.http.get<any>(this.apiUrl + '/' + id)
    }

    addMember(teamId: string, email: string, profile: string): Observable<any> {
        return this.http.patch<any>(this.apiUrl + '/' + teamId + '/members', { email, profile }).pipe(
            tap(() => {
                this.cacheService.deleteAllRelated(this.apiUrl + '/' + teamId);
            })
        );
    }

    updateMember(teamId: string, memberId: string, profile: string): Observable<any> {
        return this.http.patch<any>(this.apiUrl + '/' + teamId + '/members/' + memberId, { profile }).pipe(
            tap(() => {
                this.cacheService.deleteAllRelated(this.apiUrl + '/' + teamId);
            })
        );

    }

    removeMember(teamId: string, memberId: string): Observable<any> {
        return this.http.delete<any>(this.apiUrl + '/' + teamId + '/members/' + memberId).pipe(
            tap(() => {
                this.cacheService.deleteAllRelated(this.apiUrl);
            })
        );
    }

    getTeamspaceProjects(teamId: string): Observable<any> {
        return this.http.get<any>(this.apiUrl + '/' + teamId + '/projects')
    }

    getTeamspaceById(id: string): Observable<any> {
        return this.http.get<any>(this.apiUrl + '/' + id)
    }

    getTeamspaceClusters(teamId: string): Observable<any> {
        return this.http.get<any>(this.apiUrl + '/' + teamId + '/clusters')
    }
}



