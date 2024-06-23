export interface Cluster {
    ID: string;
    Name: string;
    Description?: string;
    Type?: string;
    Address: string;
    Port?: string;
    Token: string;
    ExpiryDate?: Date;
    CreatorID?: string;
    Teamspaces?: string[];
}
