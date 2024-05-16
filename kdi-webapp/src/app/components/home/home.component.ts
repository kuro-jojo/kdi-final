import { Component } from '@angular/core';
import { UserService } from 'src/app/_services';
import { User } from 'src/app/_interfaces/user';
import { HttpErrorResponse } from '@angular/common/http';
@Component({
    selector: 'app-home',
    templateUrl: './home.component.html',
    styleUrls: ['./home.component.css']
})
export class HomeComponent {
    id: string = '';
    user!: User;
    constructor(private userService: UserService) { }

    ngOnInit(){
        this.userService.getUserById(this.id)
        .subscribe({
            next: (resp) => {
                this.user = resp;
                console.log("User", this.user)
            },
            error: (error: HttpErrorResponse) => {
                console.error("Error getting user: ", error.error.message);
            },
            complete: () => {
                console.log("successfully");
            }
        });
    }
}
