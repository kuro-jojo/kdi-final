import { Component, ElementRef, ViewChild } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { MessageService } from 'primeng/api';
import { FileUploadHandlerEvent } from 'primeng/fileupload';
import { DeploymentService } from 'src/app/_services/deployment.service';

@Component({
    selector: 'app-add-microservice-yaml',
    templateUrl: './add-microservice-yaml.component.html',
    styleUrl: './add-microservice-yaml.component.scss'
})
export class AddMicroserviceYamlComponent {

    uploadedFiles: File[] = [];
    environmentID!: string;
    messages = {
        success: [],
        info: [],
        error: []
    };
    microservices = [];
    @ViewChild('overlay') overlay!: ElementRef;

    constructor(
        private route: ActivatedRoute,
        private deploymentService: DeploymentService,
        private messageService: MessageService
    ) { }

    ngOnInit() {
        this.route.params.subscribe(params => {
            this.environmentID = params['id'];
        });
    }

    onUpload(event: FileUploadHandlerEvent) {
        this.uploadedFiles = [];
        for (let file of event.files) {
            this.uploadedFiles.push(file);
        }
        this.overlay.nativeElement.style.display = 'block';
        this.deploymentService.addDeploymentWithYaml(this.environmentID, this.uploadedFiles).subscribe({
            next: (resp: any) => {
                this.messages = resp.messages;
                this.microservices = resp.microservices;
                this.uploadedFiles = [];
                if (this.messages && this.messages.success.length > 0) {
                    this.messageService.add({ severity: 'success', summary: 'Deployments added successfully', detail: ' ' });
                }
            },
            error: (error) => {
                this.overlay.nativeElement.style.display = 'none';
                if (error.status === 0) {
                    this.messageService.add({ severity: 'info', summary: 'Server is down', detail: 'Please try again later' });
                }
                if (error.error) {
                    this.messages = error.error.messages;
                }
                if (error.error) {
                    this.microservices = error.error.microservices;
                }
                if (this.messages && !this.messages.success || this.messages.success.length === 0) {
                    this.messageService.add({ severity: 'error', summary: 'Failed to add deployments with yaml', detail: 'Please check your yaml files' });
                }
                console.log('Error adding deployment with yaml', error);
                this.uploadedFiles = [];
            },
            complete: () => {
                this.overlay.nativeElement.style.display = 'none';
                // See if we will clear the file upload or not
            }
        });

    }
}