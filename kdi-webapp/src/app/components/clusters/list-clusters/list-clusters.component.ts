import { Component, ViewChild } from '@angular/core';
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import { MatTableDataSource } from '@angular/material/table';
import { Router } from '@angular/router';
import { MessageService } from 'primeng/api';
import { Cluster } from 'src/app/_interfaces/cluster';
import { ClusterService } from 'src/app/_services/cluster.service';
import { ReloadComponent } from 'src/app/component.util';


@Component({
    selector: 'app-list-clusters',
    templateUrl: './list-clusters.component.html',
    styleUrl: './list-clusters.component.css',
})

export class ListClustersComponent {
    displayedColumns: string[] = ['Name', 'Description', 'IpAddress', 'Port', 'CreatedAt', 'ExpiryDate',];

    dataSource: MatTableDataSource<Cluster> = new MatTableDataSource<Cluster>();
    @ViewChild(MatPaginator)
    paginator!: MatPaginator;
    @ViewChild(MatSort)
    sort!: MatSort;

    constructor(
        private clusterService: ClusterService,
        private router: Router,
        private messageService: MessageService,
    ) {
        this.clusterService.getOwnedClusters().subscribe({
            next: (resp: any) => {
                this.dataSource.data = resp.clusters as Cluster[];
                this.dataSource.paginator = this.paginator;
                this.dataSource.sort = this.sort;
            },
            error: (error) => {
                this.messageService.add({ severity: 'error', summary: "Failed to fetch clusters. Please try again later." });
                console.log(error);
            }
        });
    }

    ngAfterViewInit() {
        this.dataSource.paginator = this.paginator;
        this.dataSource.sort = this.sort;
    }
    isExpired(expiryDate: string): boolean {
        return new Date(expiryDate) < new Date();
    }

    isNearExpiry(expiryDate: string): boolean {
        const expiry = new Date(expiryDate);
        const today = new Date();
        const diff = expiry.getTime() - today.getTime();
        const days = diff / (1000 * 3600 * 24);
        return days < 3;
    }

    showClusterDetails(row: Cluster) {
        this.router.navigate(['/clusters/local/' + row.ID + '/edit']);
    }

    editCluster(row: Cluster) {
        this.router.navigate(['/clusters/' + row.ID + '/edit']);
    }

    reloadPage() {
        ReloadComponent(true, this.router);
    }
}