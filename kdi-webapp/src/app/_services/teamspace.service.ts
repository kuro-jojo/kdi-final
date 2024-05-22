import { Injectable } from '@angular/core';
import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { Observable, catchError, forkJoin, map } from 'rxjs';
import { environment } from 'src/environments/environment';
import { Teamspace } from '../_interfaces/teamspace';
import { HeadersService } from './headers.service';

@Injectable({
    providedIn: 'root'
})
export class TeamspaceService {

    readonly apiUrl = environment.apiUrl + '/dashboard/teamspaces';

    constructor(
        private http: HttpClient,
        private headerService: HeadersService
    ) { }


    createTeamspace(teamspace: Teamspace): Observable<any> {
        const headers = this.headerService.getHeaders();
        return this.http.post<any>(this.apiUrl, teamspace, { headers })
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
        return this.http.patch<any>(this.apiUrl + '/' + teamId + '/members', { email, profile })
    }

    updateMember(teamId: string, memberId: string, profile: string): Observable<any> {
        return this.http.patch<any>(this.apiUrl + '/' + teamId + '/members/' + memberId, { profile })
    }

    removeMember(teamId: string, memberId: string): Observable<any> {
        return this.http.delete<any>(this.apiUrl + '/' + teamId + '/members/' + memberId)
    }

    getTeamspaceProjects(teamId: string): Observable<any> {
        return this.http.get<any>(this.apiUrl + '/' + teamId + '/projects')
    }

    getTeamspaceById(id: string): Observable<any> {
        return this.http.get<any>(this.apiUrl + '/' + id)
    }
}