import { Component } from '@angular/core';
import { UserService } from 'src/app/_services';
import { User } from 'src/app/_interfaces/user';
import { HttpErrorResponse } from '@angular/common/http';
import { ClusterService } from 'src/app/_services/cluster.service';
import { TeamspaceService } from 'src/app/_services/teamspace.service';
import { Cluster } from 'src/app/_interfaces/cluster';
import { ProjectService } from 'src/app/_services/project.service';
@Component({
    selector: 'app-home',
    templateUrl: './home.component.html',
    styleUrls: ['./home.component.css']
})
export class HomeComponent {
    user!: User;
    numberOfClusters = 0;
    numberOfOwnedClusters = 0;
    numberOfTeamspaceClusters = 0;
    clusterFilter = 'All Clusters';

    numberOfTeamspaces = 0;
    numberOfOwnedTeamspaces = 0;
    numberOfJoinedTeamspaces = 0;
    teamspaceFilter = 'All Teamspaces';

    numberOfProjects = 0;
    numberOfOwnedProjects = 0;
    numberOfTeamspaceProject = 0;
    projectFilter = 'All Projects';

    constructor(
        private userService: UserService,
        private clusterService: ClusterService,
        private teamspaceService: TeamspaceService,
        private projectService: ProjectService,
    ) { }

    async ngOnInit() {
        try {
            await this.getUser();
            await this.getNumberOfOwnedClusters();

            await this.getNumberOfOwnedTeamspaces();

            await this.getNumberOfOwnedProjects();
            await this.getNumberOfTeamspaceProjects();

            this.numberOfClusters = this.numberOfOwnedClusters + this.numberOfTeamspaceClusters;
            this.numberOfTeamspaces = this.numberOfOwnedTeamspaces + this.numberOfJoinedTeamspaces;
            this.numberOfProjects = this.numberOfOwnedProjects + this.numberOfTeamspaceProject;
        } catch (error) {
            console.error(error);
        }
    }

    getUser(): Promise<void> {
        return new Promise((resolve, reject) => {
            this.userService.getCurrentUser().subscribe({
                next: async (resp) => {
                    this.user = resp.user;
                    this.numberOfJoinedTeamspaces = this.user.JoinedTeamspaceIDs?.length || 0;

                    // get number of clusters for each teamspace joined by the user
                    const teamspaceIDs = this.user.JoinedTeamspaceIDs || [];
                    const clusterSizePromises = teamspaceIDs.map(id => this.getTeamspaceClusters(id));
                    const clusterSizes = await Promise.all(clusterSizePromises);
                    this.numberOfTeamspaceClusters = clusterSizes.reduce((a, b) => a + b, 0);

                    resolve();
                },
                error: (error: HttpErrorResponse) => {
                    console.error(error);
                    reject(error);
                }
            });
        });
    }

    // get number of clusters for a teamspace
    getTeamspaceClusters(teamspaceID: string): Promise<number> {
        return new Promise((resolve, reject) => {
            this.teamspaceService.getTeamspaceClusters(teamspaceID).subscribe({
                next: (resp: { 'clusters': Cluster[] }) => {
                    console.log(resp.clusters.filter((t: Cluster) => t.CreatorID !== String(this.user.ID)));
                    resolve(resp.clusters.filter((t: Cluster) => t.CreatorID !== String(this.user.ID)).length);
                },
                error: (error: HttpErrorResponse) => {
                    console.error(error);
                    reject(error);
                }
            });
        });
    }

    // get number of owned clusters
    getNumberOfOwnedClusters(): Promise<void> {
        return new Promise((resolve, reject) => {
            this.clusterService.getOwnedClusters().subscribe({
                next: (resp) => {
                    this.numberOfOwnedClusters = resp.clusters? resp.clusters.length : 0;
                    resolve();
                },
                error: (error: HttpErrorResponse) => {
                    console.error(error);
                    reject(error);
                }
            });
        });
    }

    getNumberOfOwnedTeamspaces(): Promise<void> {
        return new Promise((resolve, reject) => {
            this.teamspaceService.listTeamspacesOwned().subscribe({
                next: (resp) => {
                    this.numberOfOwnedTeamspaces = resp.teamspaces ? resp.teamspaces.length : 0;
                    resolve();
                },
                error: (error: HttpErrorResponse) => {
                    console.error(error);
                    reject(error);
                }
            });
        });
    }

    getNumberOfOwnedProjects(): Promise<void> {
        return new Promise((resolve, reject) => {
            this.projectService.getOwnedProjects().subscribe({
                next: (resp) => {
                    this.numberOfOwnedProjects = resp.projects ? resp.projects.length : 0;
                    resolve();
                },
                error: (error: HttpErrorResponse) => {
                    console.error(error);
                    reject(error);
                }
            });
        });
    }

    getNumberOfTeamspaceProjects(): Promise<void> {
        return new Promise((resolve, reject) => {
            this.projectService.listProjectsOfJoinedTeamspaces().subscribe({
                next: (resp) => {
                    this.numberOfTeamspaceProject = resp.projects ? resp.projects.length : 0;
                    resolve();
                },
                error: (error: HttpErrorResponse) => {
                    console.error(error);
                    reject(error);
                }
            });
        });
    }

    filterOnCluster(filter: string) {
        switch (filter) {
            case 'o':
                this.clusterFilter = 'Owned Clusters';
                this.numberOfClusters = this.numberOfOwnedClusters;
                break;
            case 'j':
                this.clusterFilter = 'Joined Clusters';
                this.numberOfClusters = this.numberOfTeamspaceClusters;
                break;
            default:
                this.clusterFilter = 'All Clusters';
                this.numberOfClusters = this.numberOfOwnedClusters + this.numberOfTeamspaceClusters;
                break;
        }
    }

    filterOnTeamspace(filter: string) {
        switch (filter) {
            case 'o':
                this.teamspaceFilter = 'Owned Teamspaces';
                this.numberOfTeamspaces = this.numberOfOwnedTeamspaces;
                break;
            case 'j':
                this.teamspaceFilter = 'Joined Teamspaces';
                this.numberOfTeamspaces = this.numberOfJoinedTeamspaces;
                break;
            default:
                this.teamspaceFilter = 'All Teamspaces';
                this.numberOfTeamspaces = this.numberOfOwnedTeamspaces + this.numberOfJoinedTeamspaces;
                break;
        }
    }

    filterOnProject(filter: string) {
        switch (filter) {
            case 'o':
                this.projectFilter = 'Owned Projects';
                this.numberOfProjects = this.numberOfOwnedProjects;
                break;
            case 'j':
                this.projectFilter = 'Joined Projects';
                this.numberOfProjects = this.numberOfTeamspaceProject;
                break;
            default:
                this.projectFilter = 'All Projects';
                this.numberOfProjects = this.numberOfOwnedProjects + this.numberOfTeamspaceProject;
                break;
        }
    }
}