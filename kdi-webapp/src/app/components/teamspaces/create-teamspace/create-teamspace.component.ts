import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { TeamspaceService } from 'src/app/_services/teamspace.service';
import { HttpErrorResponse } from '@angular/common/http';
import { first } from 'rxjs';
import { MessageService } from 'primeng/api';

@Component({
    selector: 'app-create-teamspace',
    standalone: false,
    templateUrl: './create-teamspace.component.html',
    styleUrl: './create-teamspace.component.css'
})
export class CreateTeamspaceComponent implements OnInit {

    teamspaceForm: FormGroup;
    submitted = false;

    constructor(
        private formBuilder: FormBuilder,
        private router: Router,
        private teamspaceService: TeamspaceService,
        private messageService: MessageService,
    ) {
        this.teamspaceForm = new FormGroup({});
    }

    ngOnInit(): void {

        this.teamspaceForm = this.formBuilder.group({
            name: ['', Validators.required],
            description: ['', Validators.minLength(6)],
        });

    }

    get formControls() { return this.teamspaceForm.controls; }

    onSubmit() {
        this.submitted = true;
        // stop here if form is invalide
        if (this.teamspaceForm.invalid) {
            return;
        } this.teamspaceService.createTeamspace(this.teamspaceForm.value)
            .pipe(first())
            .subscribe({
                next: () => {
                    this.messageService.add({ severity: 'success', summary: "You have successfully created a project!" });
                    this.router.navigate(['teamspaces'])
                },
                error: (error: HttpErrorResponse) => {
                    this.messageService.add({ severity: 'error', summary: error.error.message });
                    console.error("Teamspace creation error :", error);
                }
            })
    }

}
