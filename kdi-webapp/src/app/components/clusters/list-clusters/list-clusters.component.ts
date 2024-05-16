import { HttpResponse } from '@angular/common/http';
import { Component, ViewChild } from '@angular/core';
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import { MatRow, MatTableDataSource } from '@angular/material/table';
import { Router } from '@angular/router';
import { Cluster } from 'src/app/_interfaces/cluster';
import { ClusterService } from 'src/app/_services/cluster.service';
import { ReloadComponent } from 'src/app/component.util';
import { ToastComponent } from 'src/app/components/toast/toast.component';


@Component({
    selector: 'app-list-clusters',
    templateUrl: './list-clusters.component.html',
    styleUrl: './list-clusters.component.css',
})

export class ListClustersComponent {
    @ViewChild(ToastComponent) toastComponent!: ToastComponent;
    displayedColumns: string[] = ['Name', 'Description', 'IpAddress', 'Port', 'CreatedAt', 'ExpiryDate',];

    dataSource: MatTableDataSource<Cluster> = new MatTableDataSource<Cluster>();
    @ViewChild(MatPaginator)
    paginator!: MatPaginator;
    @ViewChild(MatSort)
    sort!: MatSort;

    constructor(
        private clusterService: ClusterService,
        private router: Router
    ) {
        this.clusterService.getClusters().subscribe({
            next: (resp: any) => {
                this.dataSource.data = resp.clusters as Cluster[];
                this.dataSource.paginator = this.paginator;
                this.dataSource.sort = this.sort;
            },
            error: (error) => {
                this.toastComponent.message = "Failed to fetch clusters. Please try again later.";
                this.toastComponent.toastType = 'info';
                this.triggerToast();
                console.log(error);
            }
        });
    }

    ngAfterViewInit() {
        this.dataSource.paginator = this.paginator;
        this.dataSource.sort = this.sort;
    }
    triggerToast(): void {
        this.toastComponent.showToast();
    }

    isExpired(expiryDate: string): boolean {
        return new Date(expiryDate) < new Date();
    }

    isNearExpiry(expiryDate: string): boolean {
        const expiry = new Date(expiryDate);
        const today = new Date();
        const diff = expiry.getTime() - today.getTime();
        const days = diff / (1000 * 3600 * 24);
        return days < 7;
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