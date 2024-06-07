import { Component } from '@angular/core';
import { NotificationService } from 'src/app/_services/notification.service';
import { Notification } from 'src/app/_interfaces/notification';
import { UserService } from 'src/app/_services';
import { User } from 'src/app/_interfaces';
import { MessageService } from 'primeng/api';
import { timer } from 'rxjs';

@Component({
    selector: 'app-navbar',
    templateUrl: './navbar.component.html',
    styleUrls: ['./navbar.component.css']
})
export class NavbarComponent {

    notifications: Notification[] = [];
    unreadNotifications: number = 0;
    user: User = {} as User;

    constructor(
        private notificationService: NotificationService,
        private userService: UserService,
        private messageService: MessageService,
    ) {
    }

    ngOnInit(): void {
        this.getNotifications();
        this.getUserInfo();
    }

    getNotifications() {
        this.notificationService.getNotifications().subscribe({
            next: (resp) => {
                if (resp.notifications) {
                    this.notifications = resp.notifications;
                    this.unreadNotifications = this.notifications.filter(n => !n.WasRead).length;
                }
            },
            error: (error) => {
                console.error("Error fetching notifications: ", error);
            }
        });
    }

    getUserInfo() {
        return this.userService.getCurrentUser().subscribe({
            next: (resp) => {
                this.user = resp.user;
            },
            error: (error) => {
                console.error("Error fetching user information: ", error);
            }
        });
    }

    showNotifications($event: any) {
        $event.stopPropagation();
    }

    markAsRead(notification: Notification | '') {
        if (notification && !notification.WasRead) {
            console.log("Marking notification as read: ", notification);
            this.notificationService.readNotification(notification).subscribe({
                next: () => {
                    this.notifications = this.notifications.map(n => {
                        if (n === notification) {
                            n.WasRead = true;
                        }
                        return n;
                    })
                    this.unreadNotifications = this.notifications.filter(n => !n.WasRead).length;
                },
                error: (error) => {
                    console.error("Error reading notification: ", error);
                },
                complete: () => {
                    console.log("Notification read successfully");
                }
            });
        } else if (notification === '' && this.notifications.length > 0) {
            console.log("Marking all notifications as read");
            this.notificationService.readAllNotifications().subscribe({
                next: () => {
                    this.notifications = [];
                    this.unreadNotifications = 0;
                },
                error: (error) => {
                    console.error("Error reading all notifications: ", error);
                },
                complete: () => {
                    console.log("All notifications read successfully");
                }
            });

            // TODO : fix the display of the notifications section when all notifications are read
        }
    }

    logout() {
        if (confirm("Are you sure you want to sign out?")) {
            this.userService.logout();
            this.messageService.add({ severity: 'info', summary: 'Logged out', detail: 'You have been successfully logged out' });
            timer(1500).subscribe(() => {
                window.location.reload();
            });
        }
    }
}