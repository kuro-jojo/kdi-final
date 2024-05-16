import { Injectable } from '@angular/core';
import { HttpClient, HttpErrorResponse, HttpHeaders } from '@angular/common/http';
import { Observable, catchError, map, throwError } from 'rxjs';
import { environment } from 'src/environments/environment';
import { Project } from '../_interfaces/project';
import { HeadersService } from './headers.service';

@Injectable({
    providedIn: 'root'
})
export class ProjectService {
    readonly apiUrl = environment.apiUrl + '/dashboard/projects';

    constructor(
        private http: HttpClient,
        private headerService: HeadersService
    ) { }

    createProject(project: Project): Observable<any> {
        const headers = this.headerService.getHeaders();
        return this.http.post<any>(this.apiUrl, project, { headers })
    }

    updateProject(project: Project): Observable<any> {
        const headers = this.headerService.getHeaders();
        return this.http.patch<Project>(this.apiUrl + '/' + project.ID, project, { headers })
    }

    deleteProject(id: string): Observable<any> {
        const headers = this.headerService.getHeaders();
        return this.http.delete(this.apiUrl + '/' + id, { headers });
    }

    listProjects(): Observable<any> {
        return this.http.get<any>(this.apiUrl)
    }

    listProjectsOfJoinedTeamspaces(): Observable<any> {
        return this.http.get<any>(this.apiUrl + '/joinedTeamspaces').pipe(
            map(resp => resp),
            catchError((error: HttpErrorResponse) => {
                console.error("Error during getting projects:", error);
                throw error;
            })
        );
    }


    getProjectDetails(id: string): Observable<any> {
        return this.http.get<any>(this.apiUrl + '/' + id).pipe(
            map(resp => resp),
            catchError((error: HttpErrorResponse) => {
                console.error("Error during getting project:", error);
                throw error;
            })
        );
    }


}
