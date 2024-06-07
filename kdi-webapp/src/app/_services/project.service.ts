import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from 'src/environments/environment';
import { Project } from '../_interfaces/project';

@Injectable({
    providedIn: 'root'
})
export class ProjectService {
    readonly apiUrl = environment.apiUrl + '/dashboard/projects';

    constructor(
        private http: HttpClient,
    ) { }

    createProject(project: Project): Observable<any> {
        return this.http.post<any>(this.apiUrl, project)
    }

    updateProject(project: Project): Observable<any> {
        return this.http.patch<Project>(this.apiUrl + '/' + project.ID, project)
    }

    deleteProject(id: string): Observable<any> {
        return this.http.delete(this.apiUrl + '/' + id);
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
