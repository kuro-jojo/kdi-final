import { Injectable } from '@angular/core';

@Injectable({
    providedIn: 'root'
})
export class CacheService {
    private cache: Map<string, any[]> = new Map<string, any>()
    private localStorageName = 'httpCache'
    private cacheExpirationTime = 1000 * 60 * 5 // 5 minutes

    constructor() {
        const savedCache = localStorage.getItem(this.localStorageName);
        if (savedCache) {
            this.cache = new Map<string, any[]>(JSON.parse(savedCache));
        }
    }

    get(key: string): any[] | undefined {
        return this.cache.get(key);
    }

    set(key: string, value: any[]): void {
        this.cache.set(key, value);
        this.syncLocalStorage();
    }

    delete(key: string): void {
        this.cache.delete(key);
        this.syncLocalStorage();
    }

    deleteAllRelated(key: string): void {
        this.cache.forEach((_, k) => {
            if (k.includes(key)) {
                console.log("Deleting", k);
                this.cache.delete(k);
            }
        });
        this.syncLocalStorage();
    }
    clear(): void {
        this.cache.clear();
        localStorage.removeItem(this.localStorageName);
    }

    private syncLocalStorage(): void {
        localStorage.setItem(this.localStorageName, JSON.stringify(Array.from(this.cache.entries())));
    }

    isExpired(data: any[]): boolean {
        return data[1] < Date.now();
    }

    getExpirationTime(): number {
        return this.cacheExpirationTime;
    }
}