import { HttpClient } from "@angular/common/http";
import { Injectable } from "@angular/core";
import { Observable } from "rxjs";
import { environment } from "src/environments/environment";
import { Profile } from "../_interfaces/profile";
import { Notification } from "../_interfaces/notification";

@Injectable({
    providedIn: 'root'
})
export class NotificationService {
    readonly apiUrl = environment.apiUrl + '/dashboard/users/notifications';

    constructor(
        private http: HttpClient,
    ) { }

    getNotifications(): Observable<any> {
        return this.http.get<Notification[]>(this.apiUrl)
    }

    readNotification(notification: Notification): Observable<any> {
        return this.http.patch(this.apiUrl, notification);
    }

    readAllNotifications(): Observable<any> {
        return this.http.delete(this.apiUrl);
    }
}