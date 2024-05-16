import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { environment } from 'src/environments/environment';
import { Environment } from '../_interfaces/environment';
import { Observable, catchError, map } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class EnvironmentService {

  readonly API_URL = environment.apiUrl + '/dashboard/environments';
    constructor(
        private http: HttpClient,
    ) { }

    createEnvironment(environment: Environment ): Observable<any> {
        return this.http.post<Environment>(this.API_URL, environment)
    }

    getenvironments(): Observable<any> {
        return this.http.get<any>(this.API_URL + '/ByCluster')
    }

    getlistProjectEnvironments(id:string): Observable<any> {
      return this.http.get<any>(this.API_URL+'/projects/'+id).pipe(
        map(resp => resp),
        catchError((error: HttpErrorResponse) => {
          console.error("Error during getting environments:", error);
          throw error; 
        })
      );
    }

    getEnvironmentDetails(id: string): Observable<any> {
      return this.http.get<any>(this.API_URL+'/'+id).pipe(
        map(resp => resp), 
        catchError((error: HttpErrorResponse) => {
          console.error("Error during getting environment:", error);
          throw error; 
        })
      );
    }



  
}
