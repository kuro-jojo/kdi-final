// https://www.geeksforgeeks.org/how-to-make-a-toast-notification-in-html-css-and-javascript/

import { Component, Input, ViewEncapsulation } from '@angular/core';

@Component({
	selector: 'app-toast',
	template: '',
	styleUrls: ['./toast.component.css'],
	encapsulation: ViewEncapsulation.None // to apply styles globally
})
export class ToastComponent {
	@Input() message: string = 'Sample Message';
	@Input() toastType: 'success' | 'danger' | 'warning' | 'info' = 'info';
	@Input() duration: number = 6000;

	icons = {
		success: '<span class="material-symbols-outlined">task_alt</span>',
		danger: '<span class="material-symbols-outlined">error</span>',
		warning: '<span class="material-symbols-outlined">warning</span>',
		info: '<span class="material-symbols-outlined">info</span>'
	};

	constructor() { }

	showToast(): void {
		if (!Object.keys(this.icons).includes(this.toastType)) {
			this.toastType = 'info';
		}

		const box = document.createElement('div');
		box.classList.add('toast-container', `toast-${this.toastType}`);
		box.innerHTML = `
			<div class="toast-content-wrapper">
				<div class="toast-icon">${this.icons[this.toastType]}</div>
				<div class="toast-message">${this.message}</div>
				<div class="toast-progress"></div>
			</div>`;
		this.duration = this.duration || 5000;
		if (box != null) {
			const toastProgress = box.querySelector('.toast-progress') as HTMLElement;
			if (toastProgress != null) {
				toastProgress.style.animationDuration = `${this.duration / 1000}s`;
			}
		}

		const toastAlready = document.body.querySelector('.toast');
		if (toastAlready) {
			toastAlready.remove();
		}
		document.body.appendChild(box);
	}
}
