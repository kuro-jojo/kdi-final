import { Component, ViewChild } from '@angular/core';
import { NotificationService } from 'src/app/_services/notification.service';
import { Notification } from 'src/app/_interfaces/notification';
import { ServerService } from 'src/app/_services/server.service';
import { ToastComponent } from 'src/app/components/toast/toast.component';

@Component({
    selector: 'app-navbar',
    templateUrl: './navbar.component.html',
    styleUrls: ['./navbar.component.css']
})
export class NavbarComponent {
    @ViewChild(ToastComponent) toastComponent!: ToastComponent;
    notifications: Notification[] = [];
    unreadNotifications: number = 0;

    constructor(
        private notificationService: NotificationService,
        private serverService: ServerService,
    ) {
        this.serverService.serverStatus().subscribe({
            next: (resp) => {
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
            },
            error: () => {
                this.toastComponent.message = "Server is not available. Please try again later"
                this.toastComponent.toastType = 'info';
                this.triggerToast();
            }
        });
    }

    ngOnInit(): void {
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

    triggerToast(): void {
        this.toastComponent.showToast();
    }
}
