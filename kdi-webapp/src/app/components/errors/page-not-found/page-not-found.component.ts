import { Component } from '@angular/core';
import { Router, RouterLink } from '@angular/router';

@Component({
    selector: 'app-page-not-found',
    standalone: true,
    imports: [RouterLink],
    template: ` 
  <main class="container d-flex justify-content-center align-items-center flex-column text-center">
      <h1>404</h1>
        <h2>Ressource not found</h2>
      <div>
          <p>Sorry, the ressource you are looking for could not be found.</p>
            
      </div>

      <a class="btn" [routerLink]="['/']" >Homepage</a>
  </main>
`, styleUrl: './page-not-found.component.css'
})
export class PageNotFoundComponent {

}
