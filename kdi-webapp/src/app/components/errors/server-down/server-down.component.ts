import { Component } from '@angular/core';
import { Router } from '@angular/router';

@Component({
    selector: 'app-server-down',
    standalone: true,
    imports: [],
    template: ` 
        <main class="container d-flex justify-content-center align-items-center flex-column text-center">
            <h1>Woops! <br> Something went wrong</h1>
            <div>
                <p>Our backend is currently down .</p>
            </div>

            <button (click)="reload()">Let's try again</button>
        </main>
    `,
    styleUrl: './server-down.component.css'
})
export class ServerDownComponent {
    constructor(
        private router: Router
    ) { }

    reload() {
        this.router.navigateByUrl(window.history.state.redirect);
    }
}
