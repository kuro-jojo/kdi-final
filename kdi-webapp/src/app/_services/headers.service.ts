import { Injectable } from '@angular/core';
import { HttpHeaders } from '@angular/common/http';

@Injectable({
  providedIn: 'root'
})
export class HeadersService {

  constructor() { }

getToken(): string | null {
  return localStorage.getItem('api.accessToken');
}
getHeaders(): HttpHeaders {
  const token = this.getToken();
  if (token) {
    return new HttpHeaders({
      Authorization: `Bearer ${token}`
    });
  } else {
    return new HttpHeaders();
  }
}
}
