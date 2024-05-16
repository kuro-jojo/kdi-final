import { HttpClient } from "@angular/common/http";
import { Injectable } from "@angular/core";
import { Observable } from "rxjs";
import { environment } from "src/environments/environment";
import { Profile } from "../_interfaces/profile";

@Injectable({
    providedIn: 'root'
})
export class ProfileService {
    readonly apiUrl = environment.apiUrl + '/dashboard/profiles';

    constructor(
        private http: HttpClient,
    ) { }

    getProfiles(): Observable<any> {
        return this.http.get<Profile[]>(this.apiUrl)
    }
}