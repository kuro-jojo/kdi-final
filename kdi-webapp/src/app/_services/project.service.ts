import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, tap } from 'rxjs';
import { environment } from 'src/environments/environment';
import { Project } from '../_interfaces/project';
import { CacheService } from './cache.service';

@Injectable({
    providedIn: 'root'
})
export class ProjectService {
    readonly apiUrl = environment.apiUrl + '/dashboard/projects';

    constructor(
        private http: HttpClient,
        private cacheService: CacheService,
    ) { }

    createProject(project: Project): Observable<any> {
        return this.http.post<any>(this.apiUrl, project).pipe(
            tap(() => {
                console.log("clearing cache for " + this.apiUrl);
                this.cacheService.deleteAllRelated(this.apiUrl);
            })
        );
    }

    updateProject(project: Project): Observable<any> {
        return this.http.patch<Project>(this.apiUrl + '/' + project.ID, project).pipe(
            tap(() => {
                this.cacheService.deleteAllRelated(this.apiUrl + '/' + project.ID);
            })
        );
    }

    deleteProject(id: string): Observable<any> {
        return this.http.delete(this.apiUrl + '/' + id).pipe(
            tap(() => {
                this.cacheService.deleteAllRelated(this.apiUrl);
            })
        );
    }

    getOwnedProjects(): Observable<any> {
        return this.http.get<any>(this.apiUrl + '/owned')
    }

    listProjectsOfJoinedTeamspaces(): Observable<any> {
        return this.http.get<any>(this.apiUrl + '/joinedTeamspaces')
    }

    getProjectDetails(id: string): Observable<any> {
        return this.http.get<any>(this.apiUrl + '/' + id)
    }
}
