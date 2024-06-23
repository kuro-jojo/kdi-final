import { Time } from "@angular/common"
import { HttpEvent, HttpHandler, HttpInterceptor, HttpRequest, HttpResponse } from "@angular/common/http"
import { Injectable } from "@angular/core"
import { Observable, of, tap } from "rxjs"
import { CacheService } from "./_services/cache.service"

@Injectable()
export class CacheInterceptor implements HttpInterceptor {
    private uncacheableUrls = [
        'login',
        'register',
        'health',
    ]
    private cacheExpirationTime = 1000 * 60 * 5 // 5 minutes

    constructor(
        private cacheService: CacheService,
    ) {
    }

    intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
        if (req.method !== "GET") {
            return next.handle(req)
        }
        const key = req.urlWithParams;

        if (this.uncacheableUrls.some(url => key.includes(url))) {
            return next.handle(req)
        }

        const cachedData: any = this.cacheService.get(key);

        if (cachedData) {
            if (this.cacheService.isExpired(cachedData)) {
                this.cacheService.delete(key);
                return next.handle(req)
            }
            return of(new HttpResponse({ body: cachedData[0] }))
        } else {
            return next.handle(req).pipe(
                tap((response) => {
                    if (response instanceof HttpResponse) {
                        this.cacheService.set(key, [response.body, Date.now() + this.cacheExpirationTime]);
                    }
                }),
            );
        }
    }
}