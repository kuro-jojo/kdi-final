export interface Cluster {
    ID?: string;
    Name: string;
    Description?: string;
    IpAddress: string;
    Port?: string;
    Token: string;
    ExpiryDate?: Date;
    CreatorID?: string;
    Teamspaces?: string[];
}
