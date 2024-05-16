import { Router } from "@angular/router";

export function ReloadComponent(self: boolean, router: Router, urlToNavigateTo?: string) {
    const url = self ? router.url : urlToNavigateTo;
    router.navigateByUrl('/', { skipLocationChange: true }).then(() => {
        router.navigate([url]);
    });
}