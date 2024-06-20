import { HttpClient } from "@angular/common/http";
import { Injectable } from "@angular/core";
import { Observable, tap } from "rxjs";
import { environment } from "src/environments/environment";
import { Notification } from "../_interfaces/notification";
import { CacheService } from "./cache.service";

@Injectable({
    providedIn: 'root'
})
export class NotificationService {
    readonly apiUrl = environment.apiUrl + '/dashboard/users/notifications';

    constructor(
        private http: HttpClient,
        private cacheService: CacheService,
    ) { }

    getNotifications(): Observable<any> {
        return this.http.get<Notification[]>(this.apiUrl)
    }

    readNotification(notification: Notification): Observable<any> {
        return this.http.patch(this.apiUrl, notification).pipe(
            tap(() => {
                this.cacheService.deleteAllRelated(this.apiUrl);
            })
        );
    }

    readAllNotifications(): Observable<any> {
        return this.http.delete(this.apiUrl).pipe(
            tap(() => {
                this.cacheService.deleteAllRelated(this.apiUrl);
            })
        );
    }
}