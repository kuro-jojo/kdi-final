export interface Cluster {
    ID: string;
    Name: string;
    Description?: string;
    Type?: string;
    Address?: string;
    Port?: string;
    Token?: string;

    // Additional fields for AWS
    Region?: string;
    AccessKeyID?: string;
    SecretKey?: string;

    ExpiryDate?: Date;
    CreatorID?: string;
    Teamspaces?: string[];
}
