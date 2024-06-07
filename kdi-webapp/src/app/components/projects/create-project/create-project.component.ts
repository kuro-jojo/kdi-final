import { Component } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { ProjectService } from 'src/app/_services/project.service';
import { Router } from '@angular/router';
import { first } from 'rxjs';
import { HttpErrorResponse } from '@angular/common/http';
import { TeamspaceService } from 'src/app/_services/teamspace.service';
import { Teamspace } from 'src/app/_interfaces/teamspace';
import { MessageService } from 'primeng/api';

@Component({
    selector: 'app-create-project',
    templateUrl: './create-project.component.html',
    styleUrl: './create-project.component.css'
})
export class CreateProjectComponent {

    projectForm: FormGroup;
    submitted = false;
    teamspaces: { teamspaces: Teamspace[], size: number } = { teamspaces: [], size: 0 };

    constructor(
        private formBuilder: FormBuilder,
        private router: Router,
        private teamService: TeamspaceService,
        private projectService: ProjectService,
        private messageService: MessageService,
    ) {
        this.projectForm = new FormGroup({});
    }

    chooseTeamsapce($event: Event) {
        console.log($event.target);
        this.projectForm.controls['teamspace'].setValue($event.target ? ['value'] : '');
    }
    ngOnInit() {

        this.projectForm = this.formBuilder.group({
            name: ['', Validators.required],
            description: ['', Validators.minLength(6)],
            teamspace_id: [''],
        });
        this.teamService.listTeamspacesOwned().subscribe(
            (resp) => { this.teamspaces = resp; }
        )
    }
    get formControls() { return this.projectForm.controls; }

    onSubmit() {
        this.submitted = true;
        // stop here if form is invalide
        if (this.projectForm.invalid) {
            return;
        }

        const formData = this.projectForm.value;

        // Si Teamspace n'est pas sélectionné, retirez-le du formulaire
        if (!formData.teamspace) {
            delete formData.teamspace;
        }
        this.projectService.createProject(formData)
            .pipe(first())
            .subscribe({
                next: (resp) => {
                    this.messageService.add({ severity: 'success', summary: "You have successfully created the project!" });
                    this.router.navigate(['/projects'])
                },
                error: (error: HttpErrorResponse) => {
                    this.messageService.add({ severity: 'error', summary: error.error.message });
                    console.error("Project creation error :", error.error.message);
                }
            })
    }
}